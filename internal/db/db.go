package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err = database.Exec(schema); err != nil {
		database.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return database, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS settings (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	qb_url TEXT NOT NULL DEFAULT '',
	qb_username TEXT NOT NULL DEFAULT '',
	qb_password TEXT NOT NULL DEFAULT '',
	download_root TEXT NOT NULL DEFAULT '/downloads/anime',
	default_category TEXT NOT NULL DEFAULT 'MikanRSS',
	rss_interval INTEGER NOT NULL DEFAULT 30
);
INSERT OR IGNORE INTO settings (id) VALUES (1);
CREATE TABLE IF NOT EXISTS subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	raw_title TEXT NOT NULL,
	rss_url TEXT NOT NULL UNIQUE,
	regex TEXT NOT NULL DEFAULT '',
	save_dir_name TEXT NOT NULL,
	save_path TEXT NOT NULL,
	rule_name TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
