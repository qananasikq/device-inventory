package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
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
	return db
}

func boolPtr(v bool) *bool { return &v }

func TestCreateAndGet(t *testing.T) {
	repo := NewDeviceRepo(OpenTestDB(t))

	d := &Device{Hostname: "gw-msk-01", IP: "192.168.1.10", Location: "Moscow", IsActive: true}
	if err := repo.Create(d); err != nil {
		t.Fatalf("create: %v", err)
	}
	if d.ID == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := repo.GetByID("1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Hostname != "gw-msk-01" {
		t.Fatalf("hostname: want gw-msk-01, got %s", got.Hostname)
	}
}

func TestGetAllWithFilters(t *testing.T) {
	repo := NewDeviceRepo(OpenTestDB(t))

	repo.Create(&Device{Hostname: "sw-main-01", IP: "172.16.0.1", Location: "Moscow", IsActive: true})
	repo.Create(&Device{Hostname: "rtr-edge-03", IP: "172.16.0.2", Location: "SPB", IsActive: true})
	repo.Create(&Device{Hostname: "sw-old-02", IP: "172.16.0.3", Location: "Kazan", IsActive: false})

	devices, err := repo.GetAll(boolPtr(true), "sw")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("want 1, got %d", len(devices))
	}
	// все без фильтра
	all, _ := repo.GetAll(nil, "")
	if len(all) != 3 {
		t.Fatalf("want 3, got %d", len(all))
	}
}

func TestUpdate(t *testing.T) {
	repo := NewDeviceRepo(OpenTestDB(t))

	repo.Create(&Device{Hostname: "fw-main", IP: "10.10.0.1", Location: "Moscow", IsActive: true})

	upd, err := repo.Update("1", Device{
		Hostname: "fw-main-v2",
		IP:       "10.10.0.50",
		Location: "SPB",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Hostname != "fw-main-v2" || upd.Location != "SPB" {
		t.Fatalf("not updated: %+v", upd)
	}
}

func TestDeactivate(t *testing.T) {
	repo := NewDeviceRepo(OpenTestDB(t))

	repo.Create(&Device{Hostname: "ap-floor3", IP: "10.0.1.1", Location: "Moscow", IsActive: true})

	if err := repo.Deactivate("1"); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	got, _ := repo.GetByID("1")
	if got.IsActive {
		t.Fatal("should be inactive")
	}
}
