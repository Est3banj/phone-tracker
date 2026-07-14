package repository

import (
	"context"
	"database/sql"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type TokenRepo struct {
	db *sql.DB
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Store(ctx context.Context, token *domain.Token) error {
	query := `INSERT INTO tokens (device_id, token_hash, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, token.DeviceID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *TokenRepo) GetByDeviceID(ctx context.Context, deviceID string) (*domain.Token, error) {
	token := &domain.Token{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_id, token_hash, expires_at, created_at, revoked
		 FROM tokens WHERE device_id = ? ORDER BY created_at DESC LIMIT 1`, deviceID).
		Scan(&token.ID, &token.DeviceID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.Revoked)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return token, err
}

func (r *TokenRepo) Revoke(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tokens SET revoked = 1 WHERE id = ?`, id)
	return err
}

func (r *TokenRepo) RevokeByDevice(ctx context.Context, deviceID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tokens SET revoked = 1 WHERE device_id = ?`, deviceID)
	return err
}
