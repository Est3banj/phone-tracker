package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type CommandRepo struct {
	db *sql.DB
}

func NewCommandRepo(db *sql.DB) *CommandRepo {
	return &CommandRepo{db: db}
}

func (r *CommandRepo) Store(ctx context.Context, cmd *domain.Command) error {
	query := `INSERT INTO commands (id, device_id, action, params, status, created_at)
	           VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, cmd.ID, cmd.DeviceID, cmd.Action, cmd.Params, cmd.Status, cmd.CreatedAt)
	return err
}

func (r *CommandRepo) GetByID(ctx context.Context, id string) (*domain.Command, error) {
	cmd := &domain.Command{}
	var sentAt, ackAt, compAt sql.NullTime
	var errStr sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_id, action, params, status, created_at, sent_at, acknowledged_at, completed_at, error
		 FROM commands WHERE id = ?`, id).
		Scan(&cmd.ID, &cmd.DeviceID, &cmd.Action, &cmd.Params, &cmd.Status, &cmd.CreatedAt,
			&sentAt, &ackAt, &compAt, &errStr)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if sentAt.Valid {
			cmd.SentAt = &sentAt.Time
		}
		if ackAt.Valid {
			cmd.AcknowledgedAt = &ackAt.Time
		}
		if compAt.Valid {
			cmd.CompletedAt = &compAt.Time
		}
		if errStr.Valid {
			cmd.Error = &errStr.String
		}
	return cmd, nil
}

func (r *CommandRepo) ListByDevice(ctx context.Context, deviceID string, limit, offset int) ([]domain.Command, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_id, action, params, status, created_at, sent_at, acknowledged_at, completed_at, error
		 FROM commands WHERE device_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []domain.Command
	for rows.Next() {
		var cmd domain.Command
		var sentAt, ackAt, compAt sql.NullTime
		var errStr sql.NullString
		if err := rows.Scan(&cmd.ID, &cmd.DeviceID, &cmd.Action, &cmd.Params, &cmd.Status, &cmd.CreatedAt,
			&sentAt, &ackAt, &compAt, &errStr); err != nil {
			return nil, err
		}
		if sentAt.Valid {
			cmd.SentAt = &sentAt.Time
		}
		if ackAt.Valid {
			cmd.AcknowledgedAt = &ackAt.Time
		}
		if compAt.Valid {
			cmd.CompletedAt = &compAt.Time
		}
		if errStr.Valid {
			cmd.Error = &errStr.String
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func (r *CommandRepo) ListPending(ctx context.Context) ([]domain.Command, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_id, action, params, status, created_at
		 FROM commands WHERE status = 'pending' ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []domain.Command
	for rows.Next() {
		var cmd domain.Command
		if err := rows.Scan(&cmd.ID, &cmd.DeviceID, &cmd.Action, &cmd.Params, &cmd.Status, &cmd.CreatedAt); err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func (r *CommandRepo) UpdateStatus(ctx context.Context, id string, status domain.CommandStatus, ts time.Time) error {
	var query string
	switch status {
	case domain.CmdSent:
		query = `UPDATE commands SET status = ?, sent_at = ? WHERE id = ?`
	case domain.CmdReceived:
		query = `UPDATE commands SET status = ?, acknowledged_at = ? WHERE id = ?`
	case domain.CmdExecuted, domain.CmdFailed:
		query = `UPDATE commands SET status = ?, completed_at = ? WHERE id = ?`
	default:
		query = `UPDATE commands SET status = ? WHERE id = ?`
	}

	var err error
	if status == domain.CmdSent || status == domain.CmdReceived || status == domain.CmdExecuted || status == domain.CmdFailed {
		_, err = r.db.ExecContext(ctx, query, status, ts, id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, id)
	}
	return err
}

func (r *CommandRepo) MarkTimedOut(ctx context.Context, timeout time.Duration) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE commands SET status = 'timed_out'
		 WHERE status IN ('pending','sent') AND created_at < datetime('now', ?)`,
		"-"+timeout.String())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
