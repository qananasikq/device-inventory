package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:secret@localhost:5432/devices?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not available:", err)
	}

	// чистим таблицу перед каждым тестом
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		id          SERIAL PRIMARY KEY,
		hostname    TEXT NOT NULL,
		ip          TEXT NOT NULL,
		location    TEXT NOT NULL,
		is_active   BOOLEAN NOT NULL DEFAULT TRUE,
		created_at  TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	_, err = db.Exec("DELETE FROM devices")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// сбрасываем sequence чтобы id начинались с 1
	_, _ = db.Exec("ALTER SEQUENCE devices_id_seq RESTART WITH 1")

	t.Cleanup(func() {
		db.Exec("DELETE FROM devices")
		db.Close()
	})

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
