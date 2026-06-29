package qbittorrent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

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
