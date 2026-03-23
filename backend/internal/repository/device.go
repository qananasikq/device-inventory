package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrDeviceNotFound = errors.New("device not found")

type Device struct {
	ID        int64     `json:"id"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Location  string    `json:"location"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type DeviceRepo struct {
	db *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Create(d *Device) error {
	return r.db.QueryRow(
		`INSERT INTO devices (hostname, ip, location, is_active)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		d.Hostname, d.IP, d.Location, d.IsActive,
	).Scan(&d.ID, &d.CreatedAt)
}

// фильтры опциональные, WHERE собирается по ситуации
func (r *DeviceRepo) GetAll(isActive *bool, search string) ([]Device, error) {
	q := "SELECT id, hostname, ip, location, is_active, created_at FROM devices"
	var where []string
	var args []any
	argIdx := 1

	if isActive != nil {
		where = append(where, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *isActive)
		argIdx++
	}

	if s := strings.TrimSpace(search); s != "" {
		where = append(where, fmt.Sprintf("hostname ILIKE $%d", argIdx))
		args = append(args, "%"+s+"%")
		argIdx++
	}

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY id DESC"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Hostname, &d.IP, &d.Location, &d.IsActive, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []Device{} // nil slice -> [] в json
	}
	return out, rows.Err()
}

func (r *DeviceRepo) GetByID(id string) (*Device, error) {
	var d Device
	err := r.db.QueryRow(
		"SELECT id, hostname, ip, location, is_active, created_at FROM devices WHERE id = $1", id,
	).Scan(&d.ID, &d.Hostname, &d.IP, &d.Location, &d.IsActive, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeviceRepo) Update(id string, upd Device) (*Device, error) {
	res, err := r.db.Exec(
		"UPDATE devices SET hostname=$1, ip=$2, location=$3, is_active=$4 WHERE id=$5",
		upd.Hostname, upd.IP, upd.Location, upd.IsActive, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrDeviceNotFound
	}
	return r.GetByID(id)
}

// soft delete
func (r *DeviceRepo) Deactivate(id string) error {
	res, err := r.db.Exec("UPDATE devices SET is_active = FALSE WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrDeviceNotFound
	}
	return nil
}
