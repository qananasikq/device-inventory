package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"device-inventory/internal/handlers"
	"device-inventory/internal/repository"
)

func main() {
	db, err := repository.OpenDB()
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer db.Close()

	dev := handlers.NewDeviceHandler(repository.NewDeviceRepo(db))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("POST /devices", dev.Create)
	mux.HandleFunc("GET /devices", dev.List)
	mux.HandleFunc("GET /devices/{id}", dev.GetByID)
	mux.HandleFunc("PUT /devices/{id}", dev.Update)
	mux.HandleFunc("DELETE /devices/{id}", dev.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "*"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(origin, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server failed: %v", err)
		}
		return
	}

	// ждём пока текущие запросы завершатся
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("server close failed: %v", closeErr)
		}
	}
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
