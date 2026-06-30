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
	for _, migration := range []struct{ table, column, definition string }{
		{"subscriptions", "season", "INTEGER NOT NULL DEFAULT 1"},
		{"subscriptions", "exclude_regex", "TEXT NOT NULL DEFAULT ''"},
		{"settings", "default_exclude_regex", "TEXT NOT NULL DEFAULT ''"},
		{"settings", "latest_exclude_regex", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "bangumi_id", "INTEGER NOT NULL DEFAULT 0"},
		{"subscriptions", "broadcast_day", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "broadcast_start", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "official_url", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "bangumi_url", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "description", "TEXT NOT NULL DEFAULT ''"},
		{"subscriptions", "broadcast_day_override", "TEXT NOT NULL DEFAULT ''"},
	} {
		var exists int
		query := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name=?`, migration.table)
		if err = database.QueryRow(query, migration.column).Scan(&exists); err != nil {
			database.Close()
			return nil, fmt.Errorf("inspect database schema: %w", err)
		}
		if exists == 0 {
			query = fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, migration.table, migration.column, migration.definition)
			if _, err = database.Exec(query); err != nil {
				database.Close()
				return nil, fmt.Errorf("migrate database: %w", err)
			}
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
	,default_exclude_regex TEXT NOT NULL DEFAULT ''
	,latest_exclude_regex TEXT NOT NULL DEFAULT ''
);
INSERT OR IGNORE INTO settings (id) VALUES (1);
CREATE TABLE IF NOT EXISTS subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	raw_title TEXT NOT NULL,
	rss_url TEXT NOT NULL UNIQUE,
	regex TEXT NOT NULL DEFAULT '',
	exclude_regex TEXT NOT NULL DEFAULT '',
	save_dir_name TEXT NOT NULL,
	save_path TEXT NOT NULL,
	rule_name TEXT NOT NULL,
	bangumi_id INTEGER NOT NULL DEFAULT 0,
	broadcast_day TEXT NOT NULL DEFAULT '',
	broadcast_start TEXT NOT NULL DEFAULT '',
	official_url TEXT NOT NULL DEFAULT '',
	bangumi_url TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	broadcast_day_override TEXT NOT NULL DEFAULT '',
	season INTEGER NOT NULL DEFAULT 1,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
