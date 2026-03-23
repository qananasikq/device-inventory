package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"device-inventory/internal/repository"
)

func setup(t *testing.T) (*DeviceHandler, *repository.DeviceRepo) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hostname TEXT NOT NULL,
		ip TEXT NOT NULL,
		location TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Skip("sqlite not available:", err)
	}

	repo := repository.NewDeviceRepo(db)
	return NewDeviceHandler(repo), repo
}

func TestCreate(t *testing.T) {
	h, _ := setup(t)

	body := `{"hostname":"gw-msk-01","ip":"192.168.1.10","location":"Moscow","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/devices", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}
	var d repository.Device
	json.NewDecoder(rec.Body).Decode(&d)
	if d.ID == 0 {
		t.Fatal("expected non-zero id")
	}
}

func TestCreateInvalidIP(t *testing.T) {
	h, _ := setup(t)

	body := `{"hostname":"gw-msk-01","ip":"999.1.1.1","location":"Moscow","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/devices", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}

	var out map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out["error"] != "invalid ip address" {
		t.Fatalf("unexpected error: %+v", out)
	}
}

func TestListFilters(t *testing.T) {
	h, repo := setup(t)

	repo.Create(&repository.Device{Hostname: "sw-main-01", IP: "172.16.0.1", Location: "Moscow", IsActive: true})
	repo.Create(&repository.Device{Hostname: "rtr-edge-03", IP: "172.16.0.2", Location: "SPB", IsActive: false})

	req := httptest.NewRequest(http.MethodGet, "/devices?is_active=true&search=sw", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	var out []repository.Device
	json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 1 || out[0].Hostname != "sw-main-01" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestDelete(t *testing.T) {
	h, repo := setup(t)
	repo.Create(&repository.Device{Hostname: "sw-temp", IP: "10.10.0.7", Location: "DC1", IsActive: true})

	req := httptest.NewRequest(http.MethodDelete, "/devices/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
}
