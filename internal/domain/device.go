package domain

import "time"

type Device struct {
	DeviceID    string    `json:"device_id"`
	UserID      int64     `json:"user_id"`
	Label       string    `json:"label,omitempty"`
	TokenHash   string    `json:"-"`
	Active      bool      `json:"active"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Token struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	TokenHash  string    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	Revoked    bool      `json:"revoked"`
}
