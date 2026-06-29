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
	var seasonColumn int
	if err = database.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('subscriptions') WHERE name='season'`).Scan(&seasonColumn); err != nil {
		database.Close()
		return nil, fmt.Errorf("inspect database schema: %w", err)
	}
	if seasonColumn == 0 {
		if _, err = database.Exec(`ALTER TABLE subscriptions ADD COLUMN season INTEGER NOT NULL DEFAULT 1`); err != nil {
			database.Close()
			return nil, fmt.Errorf("migrate database: %w", err)
		}
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
	season INTEGER NOT NULL DEFAULT 1,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
