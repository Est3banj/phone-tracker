package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type LocationRepo struct {
	db *sql.DB
}

func NewLocationRepo(db *sql.DB) *LocationRepo {
	return &LocationRepo{db: db}
}

func (r *LocationRepo) Store(ctx context.Context, loc *domain.LocationReport) error {
	query := `INSERT INTO locations (device_id, latitude, longitude, altitude, accuracy, speed, battery_level, is_charging, received_at)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		loc.DeviceID, loc.Latitude, loc.Longitude, loc.Altitude,
		loc.Accuracy, loc.Speed, loc.Battery, loc.Charging, loc.ReceivedAt)
	return err
}

func (r *LocationRepo) ListByDevice(ctx context.Context, deviceID string, from, to time.Time, limit, offset int) ([]domain.LocationReport, error) {
	query := `SELECT id, device_id, latitude, longitude, altitude, accuracy, speed, battery_level, is_charging, received_at
	           FROM locations WHERE device_id = ? AND received_at >= ? AND received_at <= ?
	           ORDER BY received_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, deviceID, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.LocationReport
	for rows.Next() {
		var rpt domain.LocationReport
		err := rows.Scan(&rpt.ID, &rpt.DeviceID, &rpt.Latitude, &rpt.Longitude,
			&rpt.Altitude, &rpt.Accuracy, &rpt.Speed, &rpt.Battery, &rpt.Charging, &rpt.ReceivedAt)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rpt)
	}
	return reports, nil
}

func (r *LocationRepo) CountByDevice(ctx context.Context, deviceID string, from, to time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM locations WHERE device_id = ? AND received_at >= ? AND received_at <= ?`,
		deviceID, from, to).Scan(&count)
	return count, err
}

func (r *LocationRepo) GetLatest(ctx context.Context, deviceID string) (*domain.LocationReport, error) {
	loc := &domain.LocationReport{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_id, latitude, longitude, altitude, accuracy, speed, battery_level, is_charging, received_at
		 FROM locations WHERE device_id = ? ORDER BY received_at DESC LIMIT 1`, deviceID).
		Scan(&loc.ID, &loc.DeviceID, &loc.Latitude, &loc.Longitude,
			&loc.Altitude, &loc.Accuracy, &loc.Speed, &loc.Battery, &loc.Charging, &loc.ReceivedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return loc, err
}
