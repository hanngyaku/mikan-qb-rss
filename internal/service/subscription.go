package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/mikan-qb-rss/internal/config"
	"github.com/example/mikan-qb-rss/internal/model"
	"github.com/example/mikan-qb-rss/internal/pathutil"
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
	u, err := url.ParseRequestURI(req.RSSURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return model.Subscription{}, fmt.Errorf("invalid RSS URL")
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return model.Subscription{}, err
	}
	resp, err := s.http.Do(httpReq)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("fetch RSS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Subscription{}, fmt.Errorf("RSS returned HTTP %d", resp.StatusCode)
	}
	rawTitle, err := rss.ParseTitle(resp.Body)
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
	sub := model.Subscription{
		Name: name, RawTitle: rawTitle, RSSURL: u.String(), Regex: req.Regex,
		SaveDirName: dirName, SavePath: pathutil.Join(settings.DownloadRoot, dirName),
		RuleName: name, Enabled: true, CreatedAt: now, UpdatedAt: now,
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO subscriptions
		(name, raw_title, rss_url, regex, save_dir_name, save_path, rule_name, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		sub.Name, sub.RawTitle, sub.RSSURL, sub.Regex, sub.SaveDirName, sub.SavePath, sub.RuleName, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("save subscription: %w", err)
	}
	sub.ID, err = result.LastInsertId()
	// ponytail: qBittorrent 写入留到第二阶段，当前数据库记录就是可运行的 mock 边界。
	return sub, err
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, raw_title, rss_url, regex, save_dir_name, save_path, rule_name, enabled, created_at, updated_at FROM subscriptions ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Subscription, 0)
	for rows.Next() {
		var item model.Subscription
		if err := rows.Scan(&item.ID, &item.Name, &item.RawTitle, &item.RSSURL, &item.Regex, &item.SaveDirName, &item.SavePath, &item.RuleName, &item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
