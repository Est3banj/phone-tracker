package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Store(ctx context.Context, evt *domain.Event) error {
	query := `INSERT INTO events (device_id, event_type, payload, received_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, evt.DeviceID, evt.EventType, evt.Payload, evt.ReceivedAt)
	return err
}

func (r *EventRepo) ListByDevice(ctx context.Context, deviceID string, limit, offset int) ([]domain.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_id, event_type, payload, received_at FROM events
		 WHERE device_id = ? ORDER BY received_at DESC LIMIT ? OFFSET ?`,
		deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var evt domain.Event
		if err := rows.Scan(&evt.ID, &evt.DeviceID, &evt.EventType, &evt.Payload, &evt.ReceivedAt); err != nil {
			return nil, err
		}
		events = append(events, evt)
	}
	return events, nil
}

func (r *EventRepo) IsDuplicate(ctx context.Context, deviceID string, eventType domain.EventType, within time.Duration) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events
		 WHERE device_id = ? AND event_type = ? AND received_at > datetime('now', ?)`,
		deviceID, eventType, "-"+within.String()).Scan(&count)
	return count > 0, err
}
