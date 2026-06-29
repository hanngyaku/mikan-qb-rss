package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Rule struct {
	Enabled                   bool     `json:"enabled"`
	MustContain               string   `json:"mustContain"`
	MustNotContain            string   `json:"mustNotContain"`
	UseRegex                  bool     `json:"useRegex"`
	EpisodeFilter             string   `json:"episodeFilter"`
	SmartFilter               bool     `json:"smartFilter"`
	PreviouslyMatchedEpisodes []string `json:"previouslyMatchedEpisodes"`
	AffectedFeeds             []string `json:"affectedFeeds"`
	IgnoreDays                int      `json:"ignoreDays"`
	LastMatch                 string   `json:"lastMatch"`
	AddPaused                 bool     `json:"addPaused"`
	AssignedCategory          string   `json:"assignedCategory"`
	SavePath                  string   `json:"savePath"`
}

type Torrent struct {
	Hash     string  `json:"hash"`
	Progress float64 `json:"progress"`
	SavePath string  `json:"save_path"`
}

type TorrentFile struct {
	Name string `json:"name"`
}

type Client struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

func New(baseURL, username, password string) (*Client, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("invalid qBittorrent URL")
	}
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"), username: username, password: password,
		http: &http.Client{Jar: jar, Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) Login(ctx context.Context) error {
	form := url.Values{"username": {c.username}, "password": {c.password}}
	body, err := c.request(ctx, http.MethodPost, "/api/v2/auth/login", strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}
	if strings.TrimSpace(body) != "Ok." {
		return fmt.Errorf("qBittorrent login failed")
	}
	return nil
}

func (c *Client) Version(ctx context.Context) (string, error) {
	return c.request(ctx, http.MethodGet, "/api/v2/app/version", nil, "")
}

func (c *Client) WebAPIVersion(ctx context.Context) (string, error) {
	return c.request(ctx, http.MethodGet, "/api/v2/app/webapiVersion", nil, "")
}

func (c *Client) Test(ctx context.Context) (string, string, error) {
	if err := c.Login(ctx); err != nil {
		return "", "", err
	}
	version, err := c.Version(ctx)
	if err != nil {
		return "", "", err
	}
	apiVersion, err := c.WebAPIVersion(ctx)
	return strings.TrimSpace(version), strings.TrimSpace(apiVersion), err
}

func (c *Client) Categories(ctx context.Context) (map[string]json.RawMessage, error) {
	body, err := c.request(ctx, http.MethodGet, "/api/v2/torrents/categories", nil, "")
	if err != nil {
		return nil, err
	}
	var categories map[string]json.RawMessage
	err = json.Unmarshal([]byte(body), &categories)
	return categories, err
}

func (c *Client) EnsureCategory(ctx context.Context, category string) error {
	categories, err := c.Categories(ctx)
	if err != nil {
		return err
	}
	if _, exists := categories[category]; exists {
		return nil
	}
	return c.postForm(ctx, "/api/v2/torrents/createCategory", url.Values{"category": {category}, "savePath": {""}})
}

func (c *Client) AddFeed(ctx context.Context, feedURL, feedPath string) error {
	return c.postForm(ctx, "/api/v2/rss/addFeed", url.Values{"url": {feedURL}, "path": {feedPath}})
}

func (c *Client) RSSItems(ctx context.Context) (json.RawMessage, error) {
	body, err := c.request(ctx, http.MethodGet, "/api/v2/rss/items?withData=false", nil, "")
	return json.RawMessage(body), err
}

func (c *Client) FeedURL(ctx context.Context, name string) (string, bool, error) {
	body, err := c.RSSItems(ctx)
	if err != nil {
		return "", false, err
	}
	var items map[string]struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &items); err != nil {
		return "", false, err
	}
	item, exists := items[name]
	return item.URL, exists, nil
}

func (c *Client) RSSRules(ctx context.Context) (json.RawMessage, error) {
	body, err := c.request(ctx, http.MethodGet, "/api/v2/rss/rules", nil, "")
	return json.RawMessage(body), err
}

func (c *Client) RSSRule(ctx context.Context, name string) (Rule, bool, error) {
	body, err := c.RSSRules(ctx)
	if err != nil {
		return Rule{}, false, err
	}
	var rules map[string]Rule
	if err := json.Unmarshal(body, &rules); err != nil {
		return Rule{}, false, err
	}
	rule, exists := rules[name]
	return rule, exists, nil
}

func (c *Client) RemoveFeed(ctx context.Context, feedPath string) error {
	return c.postForm(ctx, "/api/v2/rss/removeItem", url.Values{"path": {feedPath}})
}

func (c *Client) SetRule(ctx context.Context, name string, rule Rule) error {
	definition, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	return c.postForm(ctx, "/api/v2/rss/setRule", url.Values{"ruleName": {name}, "ruleDef": {string(definition)}})
}

func (c *Client) RemoveRule(ctx context.Context, name string) error {
	return c.postForm(ctx, "/api/v2/rss/removeRule", url.Values{"ruleName": {name}})
}

func (c *Client) Torrents(ctx context.Context, category string) ([]Torrent, error) {
	body, err := c.request(ctx, http.MethodGet, "/api/v2/torrents/info?category="+url.QueryEscape(category), nil, "")
	if err != nil {
		return nil, err
	}
	var torrents []Torrent
	err = json.Unmarshal([]byte(body), &torrents)
	return torrents, err
}

func (c *Client) TorrentFiles(ctx context.Context, hash string) ([]TorrentFile, error) {
	body, err := c.request(ctx, http.MethodGet, "/api/v2/torrents/files?hash="+url.QueryEscape(hash), nil, "")
	if err != nil {
		return nil, err
	}
	var files []TorrentFile
	err = json.Unmarshal([]byte(body), &files)
	return files, err
}

func (c *Client) RenameFile(ctx context.Context, hash, oldPath, newPath string) error {
	return c.postForm(ctx, "/api/v2/torrents/renameFile", url.Values{
		"hash": {hash}, "oldPath": {oldPath}, "newPath": {newPath},
	})
}

func (c *Client) postForm(ctx context.Context, path string, form url.Values) error {
	_, err := c.request(ctx, http.MethodPost, path, strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
	return err
}

func (c *Client) request(ctx context.Context, method, path string, body io.Reader, contentType string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return "", err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("qBittorrent request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("qBittorrent returned HTTP %d", resp.StatusCode)
	}
	return string(data), nil
}
