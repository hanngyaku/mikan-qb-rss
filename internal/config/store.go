package config

import (
	"context"
	"database/sql"

	"github.com/example/mikan-qb-rss/internal/model"
)

func Get(ctx context.Context, db *sql.DB) (model.Settings, error) {
	var s model.Settings
	err := db.QueryRowContext(ctx, `SELECT qb_url, qb_username, qb_password, download_root, default_category, rss_interval, default_exclude_regex, latest_exclude_regex FROM settings WHERE id=1`).
		Scan(&s.QBURL, &s.QBUsername, &s.QBPassword, &s.DownloadRoot, &s.DefaultCategory, &s.RSSInterval, &s.DefaultExcludeRegex, &s.LatestExcludeRegex)
	return s, err
}

func Update(ctx context.Context, db *sql.DB, req model.UpdateSettingsRequest) error {
	if req.QBPassword == "" {
		_, err := db.ExecContext(ctx, `UPDATE settings SET qb_url=?, qb_username=?, download_root=?, default_category=?, rss_interval=?, default_exclude_regex=? WHERE id=1`,
			req.QBURL, req.QBUsername, req.DownloadRoot, req.DefaultCategory, req.RSSInterval, req.DefaultExcludeRegex)
		return err
	}
	_, err := db.ExecContext(ctx, `UPDATE settings SET qb_url=?, qb_username=?, qb_password=?, download_root=?, default_category=?, rss_interval=?, default_exclude_regex=? WHERE id=1`,
		req.QBURL, req.QBUsername, req.QBPassword, req.DownloadRoot, req.DefaultCategory, req.RSSInterval, req.DefaultExcludeRegex)
	return err
}

func Public(s model.Settings) model.SettingsResponse {
	return model.SettingsResponse{
		QBURL: s.QBURL, QBUsername: s.QBUsername, PasswordSet: s.QBPassword != "",
		DownloadRoot: s.DownloadRoot, DefaultCategory: s.DefaultCategory, RSSInterval: s.RSSInterval,
		DefaultExcludeRegex: s.DefaultExcludeRegex, LatestExcludeRegex: s.LatestExcludeRegex,
	}
}

func SetLatestExcludeRegex(ctx context.Context, db *sql.DB, value string) error {
	_, err := db.ExecContext(ctx, `UPDATE settings SET latest_exclude_regex=? WHERE id=1`, value)
	return err
}
