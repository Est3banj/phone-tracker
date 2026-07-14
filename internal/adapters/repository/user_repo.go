package repository

import (
	"context"
	"database/sql"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Store(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, password_hash, role, active) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.Role, user.Active)
	return err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, active, created_at, updated_at
		 FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, active, created_at, updated_at
		 FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, password_hash, role, active, created_at, updated_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) UpdateRole(ctx context.Context, id int64, role domain.Role) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, role, id)
	return err
}

func (r *UserRepo) SetActive(ctx context.Context, id int64, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, active, id)
	return err
}
