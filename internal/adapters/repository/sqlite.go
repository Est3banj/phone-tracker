package repository

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS devices (
		device_id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id),
		label TEXT,
		token_hash TEXT NOT NULL,
		active INTEGER NOT NULL DEFAULT 1,
		last_seen DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS locations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL REFERENCES devices(device_id),
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		altitude REAL,
		accuracy REAL,
		speed REAL,
		battery_level INTEGER NOT NULL,
		is_charging INTEGER NOT NULL DEFAULT 0,
		received_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_locations_device_time ON locations(device_id, received_at);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL REFERENCES devices(device_id),
		event_type TEXT NOT NULL,
		payload TEXT,
		received_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_events_device_time ON events(device_id, received_at);

	CREATE TABLE IF NOT EXISTS commands (
		id TEXT PRIMARY KEY,
		device_id TEXT NOT NULL REFERENCES devices(device_id),
		action TEXT NOT NULL,
		params TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		sent_at DATETIME,
		acknowledged_at DATETIME,
		completed_at DATETIME,
		error TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_commands_device ON commands(device_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);

	CREATE TABLE IF NOT EXISTS tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL REFERENCES devices(device_id),
		token_hash TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		revoked INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Migration error: %v", err)
		return err
	}
	return nil
}
