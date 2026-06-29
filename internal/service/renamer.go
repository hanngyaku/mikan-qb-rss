package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/example/mikan-qb-rss/internal/config"
	"github.com/example/mikan-qb-rss/internal/model"
	"github.com/example/mikan-qb-rss/internal/qbittorrent"
)

var episodePattern = regexp.MustCompile(`(?:\[(\d{1,3})(?:v\d+)?\]| - (\d{1,3})(?:v\d+)?(?: |$))`)

type Renamer struct {
	db *sql.DB
}

func NewRenamer(db *sql.DB) *Renamer {
	return &Renamer{db: db}
}

func (r *Renamer) Start(ctx context.Context) {
	go func() {
		// ponytail: 固定一分钟轮询；任务量大时再改为可配置或完成事件回调。
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			if err := r.Run(ctx); err != nil {
				log.Printf("rename scan: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (r *Renamer) Run(ctx context.Context) error {
	settings, err := config.Get(ctx, r.db)
	if err != nil || settings.QBURL == "" {
		return err
	}
	client, err := qbittorrent.New(settings.QBURL, settings.QBUsername, settings.QBPassword)
	if err != nil {
		return err
	}
	if err := client.Login(ctx); err != nil {
		return err
	}
	torrents, err := client.Torrents(ctx, settings.DefaultCategory)
	if err != nil {
		return err
	}
	subscriptions, err := NewSubscriptionService(r.db).List(ctx)
	if err != nil {
		return err
	}
	for _, torrent := range torrents {
		if torrent.Progress < 1 {
			continue
		}
		for _, subscription := range subscriptions {
			if subscription.Enabled && samePath(torrent.SavePath, subscription.SavePath) {
				if err := renameTorrent(ctx, client, torrent, subscription); err != nil {
					log.Printf("rename torrent %s: %v", torrent.Hash, err)
				}
				break
			}
		}
	}
	return nil
}

func renameTorrent(ctx context.Context, client *qbittorrent.Client, torrent qbittorrent.Torrent, subscription model.Subscription) error {
	files, err := client.TorrentFiles(ctx, torrent.Hash)
	if err != nil {
		return err
	}
	for _, file := range files {
		episode, ok := episodeNumber(file.Name)
		if !ok || !isVideo(file.Name) {
			continue
		}
		ext := filepath.Ext(file.Name)
		newName := fmt.Sprintf("%s S%02d E%02d%s", subscription.Name, subscription.Season, episode, ext)
		newPath := path.Join(path.Dir(strings.ReplaceAll(file.Name, `\`, "/")), newName)
		if file.Name == newPath || strings.HasSuffix(strings.ReplaceAll(file.Name, `\`, "/"), "/"+newName) {
			continue
		}
		if err := client.RenameFile(ctx, torrent.Hash, file.Name, newPath); err != nil {
			return err
		}
	}
	return nil
}

func episodeNumber(name string) (int, bool) {
	match := episodePattern.FindStringSubmatch(filepath.Base(name))
	if len(match) != 3 {
		return 0, false
	}
	value := match[1]
	if value == "" {
		value = match[2]
	}
	episode, err := strconv.Atoi(value)
	return episode, err == nil
}

func isVideo(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".m4v":
		return true
	default:
		return false
	}
}

func samePath(a, b string) bool {
	normalize := func(value string) string {
		return strings.TrimRight(strings.ToLower(strings.ReplaceAll(value, `\`, "/")), "/")
	}
	return normalize(a) == normalize(b)
}
