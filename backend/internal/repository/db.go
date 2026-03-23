package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// OpenDB подключается к PostgreSQL и создаёт таблицы/индексы при старте
func OpenDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:secret@localhost:5432/devices?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		id          SERIAL PRIMARY KEY,
		hostname    TEXT NOT NULL,
		ip          TEXT NOT NULL,
		location    TEXT NOT NULL,
		is_active   BOOLEAN NOT NULL DEFAULT TRUE,
		created_at  TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("create devices table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS configs (
		id          SERIAL PRIMARY KEY,
		device_id   INTEGER NOT NULL REFERENCES devices(id),
		config_text TEXT NOT NULL,
		version     TEXT,
		created_at  TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("create configs table: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		id          SERIAL PRIMARY KEY,
		device_id   INTEGER NOT NULL REFERENCES devices(id),
		level       TEXT NOT NULL DEFAULT 'info',
		message     TEXT NOT NULL,
		created_at  TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("create logs table: %w", err)
	}

	// индексы (IF NOT EXISTS поддерживается в PostgreSQL 9.5+)
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS devices_active ON devices(is_active)`)
	if err != nil {
		return fmt.Errorf("create devices index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS configs_device ON configs(device_id)`)
	if err != nil {
		return fmt.Errorf("create configs index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS logs_device_ts ON logs(device_id, created_at DESC)`)
	if err != nil {
		return fmt.Errorf("create logs index: %w", err)
	}

	return nil
}
