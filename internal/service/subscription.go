package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/example/mikan-qb-rss/internal/config"
	"github.com/example/mikan-qb-rss/internal/model"
	"github.com/example/mikan-qb-rss/internal/pathutil"
	"github.com/example/mikan-qb-rss/internal/qbittorrent"
	"github.com/example/mikan-qb-rss/internal/rss"
)

type SubscriptionService struct {
	db   *sql.DB
	http *http.Client
}

func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{db: db, http: &http.Client{Timeout: 15 * time.Second}}
}

func (s *SubscriptionService) Create(ctx context.Context, req model.CreateSubscriptionRequest) (model.Subscription, error) {
	rssURL, err := validURL(req.RSSURL)
	if err != nil {
		return model.Subscription{}, err
	}
	rawTitle, err := s.fetchTitle(ctx, rssURL)
	if err != nil {
		return model.Subscription{}, err
	}
	name := rss.AnimeName(rawTitle)
	dirName := pathutil.CleanDirName(name)
	if strings.TrimSpace(req.CustomDirName) != "" {
		dirName = pathutil.CleanDirName(req.CustomDirName)
	}
	settings, err := config.Get(ctx, s.db)
	if err != nil {
		return model.Subscription{}, err
	}
	now := time.Now().UTC()
	season := max(req.Season, 1)
	item := model.Subscription{
		Name: name, RawTitle: rawTitle, RSSURL: rssURL, Regex: req.Regex,
		SaveDirName: dirName, Season: season,
		SavePath: pathutil.Join(pathutil.Join(settings.DownloadRoot, dirName), fmt.Sprintf("Season %d", season)),
		RuleName: name, Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.syncQB(ctx, item); err != nil {
		return model.Subscription{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO subscriptions
		(name, raw_title, rss_url, regex, save_dir_name, save_path, rule_name, season, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		item.Name, item.RawTitle, item.RSSURL, item.Regex, item.SaveDirName, item.SavePath, item.RuleName, item.Season, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		_ = s.removeQB(ctx, item)
		return model.Subscription{}, fmt.Errorf("save subscription: %w", err)
	}
	item.ID, err = result.LastInsertId()
	return item, err
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, raw_title, rss_url, regex, save_dir_name, save_path, rule_name, season, enabled, created_at, updated_at FROM subscriptions ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Subscription, 0)
	for rows.Next() {
		var item model.Subscription
		if err := rows.Scan(&item.ID, &item.Name, &item.RawTitle, &item.RSSURL, &item.Regex, &item.SaveDirName, &item.SavePath, &item.RuleName, &item.Season, &item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SubscriptionService) Get(ctx context.Context, id int64) (model.Subscription, error) {
	var item model.Subscription
	row := s.db.QueryRowContext(ctx, `SELECT id, name, raw_title, rss_url, regex, save_dir_name, save_path, rule_name, season, enabled, created_at, updated_at FROM subscriptions WHERE id=?`, id)
	err := row.Scan(&item.ID, &item.Name, &item.RawTitle, &item.RSSURL, &item.Regex, &item.SaveDirName, &item.SavePath, &item.RuleName, &item.Season, &item.Enabled, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (s *SubscriptionService) Update(ctx context.Context, id int64, req model.UpdateSubscriptionRequest) (model.Subscription, error) {
	item, err := s.Get(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}
	item.RSSURL, err = validURL(req.RSSURL)
	if err != nil {
		return model.Subscription{}, err
	}
	settings, err := config.Get(ctx, s.db)
	if err != nil {
		return model.Subscription{}, err
	}
	item.Regex = req.Regex
	item.SaveDirName = pathutil.CleanDirName(req.SaveDirName)
	item.Season = max(req.Season, 1)
	item.SavePath = pathutil.Join(pathutil.Join(settings.DownloadRoot, item.SaveDirName), fmt.Sprintf("Season %d", item.Season))
	item.Enabled = req.Enabled
	item.UpdatedAt = time.Now().UTC()
	if err := s.syncQB(ctx, item); err != nil {
		return model.Subscription{}, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE subscriptions SET rss_url=?, regex=?, save_dir_name=?, save_path=?, season=?, enabled=?, updated_at=? WHERE id=?`,
		item.RSSURL, item.Regex, item.SaveDirName, item.SavePath, item.Season, item.Enabled, item.UpdatedAt, item.ID)
	return item, err
}

func (s *SubscriptionService) Sync(ctx context.Context, id int64) (model.Subscription, error) {
	item, err := s.Get(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}
	err = s.syncQB(ctx, item)
	return item, err
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64) error {
	item, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.removeQB(ctx, item); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id=?`, id)
	return err
}

func (s *SubscriptionService) syncQB(ctx context.Context, item model.Subscription) error {
	settings, client, err := s.qbClient(ctx)
	if err != nil {
		return err
	}
	if err := client.EnsureCategory(ctx, settings.DefaultCategory); err != nil {
		return fmt.Errorf("ensure qBittorrent category: %w", err)
	}
	feedURL, feedExists, err := client.FeedURL(ctx, item.Name)
	if err != nil {
		return fmt.Errorf("get qBittorrent RSS feeds: %w", err)
	}
	if feedExists && feedURL != item.RSSURL {
		if err := client.RemoveFeed(ctx, item.Name); err != nil {
			return fmt.Errorf("replace qBittorrent RSS feed: %w", err)
		}
		feedExists = false
	}
	if !feedExists {
		if err := client.AddFeed(ctx, item.RSSURL, item.Name); err != nil {
			return fmt.Errorf("add qBittorrent RSS feed: %w", err)
		}
	}
	rule := qbittorrent.Rule{
		Enabled: item.Enabled, MustContain: item.Regex, UseRegex: item.Regex != "",
		PreviouslyMatchedEpisodes: []string{}, AffectedFeeds: []string{item.RSSURL},
		AssignedCategory: settings.DefaultCategory, SavePath: item.SavePath,
	}
	existing, ruleExists, err := client.RSSRule(ctx, item.RuleName)
	if err != nil {
		return fmt.Errorf("get qBittorrent RSS rules: %w", err)
	}
	if ruleExists {
		rule.PreviouslyMatchedEpisodes = existing.PreviouslyMatchedEpisodes
		rule.LastMatch = existing.LastMatch
		if sameRule(existing, rule) {
			return nil
		}
	}
	if err := client.SetRule(ctx, item.RuleName, rule); err != nil {
		return fmt.Errorf("set qBittorrent RSS rule: %w", err)
	}
	return nil
}

func sameRule(a, b qbittorrent.Rule) bool {
	return a.Enabled == b.Enabled &&
		a.MustContain == b.MustContain &&
		a.UseRegex == b.UseRegex &&
		a.AssignedCategory == b.AssignedCategory &&
		a.SavePath == b.SavePath &&
		slices.Equal(a.AffectedFeeds, b.AffectedFeeds)
}

func (s *SubscriptionService) removeQB(ctx context.Context, item model.Subscription) error {
	_, client, err := s.qbClient(ctx)
	if err != nil {
		return err
	}
	if _, exists, err := client.RSSRule(ctx, item.RuleName); err != nil {
		return err
	} else if exists {
		if err := client.RemoveRule(ctx, item.RuleName); err != nil {
			return fmt.Errorf("remove qBittorrent rule: %w", err)
		}
	}
	feedPath, exists, err := client.FeedPathByURL(ctx, item.RSSURL)
	if err != nil {
		return err
	}
	if exists {
		if err := client.RemoveFeed(ctx, feedPath); err != nil {
			return fmt.Errorf("remove qBittorrent feed: %w", err)
		}
	}
	return nil
}

func (s *SubscriptionService) qbClient(ctx context.Context) (model.Settings, *qbittorrent.Client, error) {
	settings, err := config.Get(ctx, s.db)
	if err != nil {
		return settings, nil, err
	}
	client, err := qbittorrent.New(settings.QBURL, settings.QBUsername, settings.QBPassword)
	if err != nil {
		return settings, nil, err
	}
	if err := client.Login(ctx); err != nil {
		return settings, nil, err
	}
	return settings, client, nil
}

func (s *SubscriptionService) fetchTitle(ctx context.Context, rssURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rssURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch RSS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("RSS returned HTTP %d", resp.StatusCode)
	}
	return rss.ParseTitle(resp.Body)
}

func validURL(value string) (string, error) {
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("invalid RSS URL")
	}
	return u.String(), nil
}
