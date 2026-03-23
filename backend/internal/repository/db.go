package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// создаём таблицы и индексы при старте
func OpenDB() (*sql.DB, error) {
	dsn := os.Getenv("DB_CONNECTION")
	if dsn == "" {
		dsn = "device_inventory.db"
	}

	if dir := filepath.Dir(dsn); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hostname TEXT NOT NULL,
		ip TEXT NOT NULL,
		location TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id INTEGER NOT NULL,
		config_text TEXT NOT NULL,
		version TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (device_id) REFERENCES devices(id)
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create configs table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id INTEGER NOT NULL,
		level TEXT NOT NULL DEFAULT 'info',
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (device_id) REFERENCES devices(id)
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create logs table: %w", err)
	}

	// индексы
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS devices_active ON devices(is_active)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create devices index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS configs_device ON configs(device_id)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create configs index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS logs_device_ts ON logs(device_id, created_at DESC)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create logs index: %w", err)
	}

	return db, nil
}
