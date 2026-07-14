package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type DeviceRepo struct {
	db *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Store(ctx context.Context, device *domain.Device) error {
	query := `INSERT INTO devices (device_id, user_id, label, token_hash, active) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, device.DeviceID, device.UserID, device.Label, device.TokenHash, device.Active)
	return err
}

func (r *DeviceRepo) GetByID(ctx context.Context, deviceID string) (*domain.Device, error) {
	dev := &domain.Device{}
	var lastSeen sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT device_id, user_id, label, token_hash, active, last_seen, created_at
		 FROM devices WHERE device_id = ?`, deviceID).
		Scan(&dev.DeviceID, &dev.UserID, &dev.Label, &dev.TokenHash, &dev.Active, &lastSeen, &dev.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if lastSeen.Valid {
		dev.LastSeen = &lastSeen.Time
	}
	return dev, err
}

func (r *DeviceRepo) ListByUser(ctx context.Context, userID int64) ([]domain.Device, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT device_id, user_id, label, token_hash, active, last_seen, created_at
		 FROM devices WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var dev domain.Device
		var lastSeen sql.NullTime
		if err := rows.Scan(&dev.DeviceID, &dev.UserID, &dev.Label, &dev.TokenHash, &dev.Active, &lastSeen, &dev.CreatedAt); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			dev.LastSeen = &lastSeen.Time
		}
		devices = append(devices, dev)
	}
	return devices, nil
}

func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, deviceID string, ts time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET last_seen = ? WHERE device_id = ?`, ts, deviceID)
	return err
}

func (r *DeviceRepo) SetActive(ctx context.Context, deviceID string, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET active = ? WHERE device_id = ?`, active, deviceID)
	return err
}
