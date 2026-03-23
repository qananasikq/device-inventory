package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

// DeviceHandler хэндлер с зависимостью от БД, чтобы не тащить глобальную переменную
type DeviceHandler struct {
	DB *sql.DB
}

// Device  модель устройства.
type Device struct {
	ID       int64
	Hostname string
	IP       string
}

// initDB открывает соединение и сразу пингует базу чтобы упасть при старте а не потом
func initDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "host=localhost user=app dbname=devices sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// sql.Open только валидирует DSN, реальное соединение через Ping
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}

// ServeHTTP обрабатывает GET /device?id=...
func (h *DeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// наследуем контекст запроса чтобы при дисконнекте клиента запрос к БД тоже отменился
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	// убрал горутину с time.Sleep по одной на каждый запрос

	const query = "SELECT id, hostname, ip FROM devices WHERE id = $1"
	row := h.DB.QueryRowContext(ctx, query, idStr)

	var d Device
	err := row.Scan(&d.ID, &d.Hostname, &d.IP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "device not found", http.StatusNotFound)
		} else {
			log.Printf("scan error: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
		}
		return
	}

	_, err = h.DB.ExecContext(ctx,
		"INSERT INTO audit_log(device_id, ts, action) VALUES ($1, now(), 'view')", d.ID)
	if err != nil {
		// аудит вторичен  не роняем запрос но логируем
		log.Printf("audit_log insert error: %v", err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Device: %s (%s)", d.Hostname, d.IP)
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("failed to init DB: %v", err)
	}
	defer db.Close()

	handler := &DeviceHandler{DB: db}
	http.Handle("/device", handler)

	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
