package service

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/example/mikan-qb-rss/internal/model"
)

var mikanPosterPattern = regexp.MustCompile(`bangumi-poster[^>]+background-image:\s*url\(['"]?([^'")]+)`)
var htmlTagPattern = regexp.MustCompile(`(?s)<[^>]+>`)

type mikanMetadata struct {
	BroadcastDay, BroadcastStart, OfficialURL, BangumiURL, Description string
}

func mikanBangumiID(rssURL string) int {
	u, err := url.Parse(rssURL)
	if err != nil || !strings.EqualFold(u.Hostname(), "mikanani.me") {
		return 0
	}
	id, _ := strconv.Atoi(u.Query().Get("bangumiId"))
	return id
}

func (s *SubscriptionService) fetchMikanMetadata(ctx context.Context, id int) (mikanMetadata, error) {
	pageURL := fmt.Sprintf("https://mikanani.me/Home/Bangumi/%d", id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	req.Header.Set("User-Agent", "mikan-qb-rss/0.1")
	resp, err := s.http.Do(req)
	if err != nil {
		return mikanMetadata{}, fmt.Errorf("fetch Mikan metadata: %w", err)
	}
	defer resp.Body.Close()
	page, err := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		return mikanMetadata{}, fmt.Errorf("fetch Mikan metadata: HTTP %d", resp.StatusCode)
	}
	metadata := mikanMetadata{
		BroadcastDay:   pageText(page, `放送日期：(.*?)</p>`),
		BroadcastStart: pageText(page, `放送开始：(.*?)</p>`),
		OfficialURL:    pageText(page, `官方网站：.*?href="([^"]+)"`),
		BangumiURL:     pageText(page, `Bangumi番组计划链接：.*?href="([^"]+)"`),
		Description:    pageText(page, `<p class="header2-desc">(.*?)</p>`),
	}
	match := mikanPosterPattern.FindSubmatch(page)
	if len(match) != 2 {
		return mikanMetadata{}, fmt.Errorf("Mikan poster not found")
	}
	if s.dataDir == "" {
		return metadata, nil
	}
	dir := filepath.Join(s.dataDir, "posters")
	target := filepath.Join(dir, strconv.Itoa(id)+".webp")
	if _, err := os.Stat(target); err == nil {
		return metadata, nil
	}
	base, _ := url.Parse(pageURL)
	imageRef, err := url.Parse(string(match[1]))
	if err != nil {
		return mikanMetadata{}, err
	}
	imageReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, base.ResolveReference(imageRef).String(), nil)
	imageReq.Header.Set("User-Agent", "mikan-qb-rss/0.1")
	imageResp, err := s.http.Do(imageReq)
	if err != nil {
		return mikanMetadata{}, fmt.Errorf("download Mikan poster: %w", err)
	}
	defer imageResp.Body.Close()
	if imageResp.StatusCode != http.StatusOK {
		return mikanMetadata{}, fmt.Errorf("download Mikan poster: HTTP %d", imageResp.StatusCode)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mikanMetadata{}, err
	}
	file, err := os.CreateTemp(dir, "poster-*")
	if err != nil {
		return mikanMetadata{}, err
	}
	tempName := file.Name()
	defer os.Remove(tempName)
	_, copyErr := io.Copy(file, io.LimitReader(imageResp.Body, 10<<20))
	closeErr := file.Close()
	if copyErr != nil {
		return mikanMetadata{}, copyErr
	}
	if closeErr != nil {
		return mikanMetadata{}, closeErr
	}
	return metadata, os.Rename(tempName, target)
}

func pageText(page []byte, pattern string) string {
	match := regexp.MustCompile(`(?s)` + pattern).FindSubmatch(page)
	if len(match) != 2 {
		return ""
	}
	value := strings.ReplaceAll(string(match[1]), "<br />", "\n")
	value = htmlTagPattern.ReplaceAllString(value, "")
	return strings.TrimSpace(html.UnescapeString(value))
}

func applyMikanMetadata(item *model.Subscription, metadata mikanMetadata) {
	item.MetadataBroadcastDay = metadata.BroadcastDay
	item.BroadcastDay = metadata.BroadcastDay
	item.BroadcastStart = metadata.BroadcastStart
	item.OfficialURL = metadata.OfficialURL
	item.BangumiURL = metadata.BangumiURL
	item.Description = metadata.Description
}

func (s *SubscriptionService) decorate(item *model.Subscription) {
	if item.BangumiID < 1 || s.dataDir == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(s.dataDir, "posters", strconv.Itoa(item.BangumiID)+".webp")); err == nil {
		item.PosterURL = fmt.Sprintf("/api/posters/%d", item.BangumiID)
	}
}
