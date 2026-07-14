package ports

import (
	"context"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
)

type LocationRepository interface {
	Store(ctx context.Context, loc *domain.LocationReport) error
	ListByDevice(ctx context.Context, deviceID string, from, to time.Time, limit, offset int) ([]domain.LocationReport, error)
	CountByDevice(ctx context.Context, deviceID string, from, to time.Time) (int, error)
	GetLatest(ctx context.Context, deviceID string) (*domain.LocationReport, error)
}

type EventRepository interface {
	Store(ctx context.Context, evt *domain.Event) error
	ListByDevice(ctx context.Context, deviceID string, limit, offset int) ([]domain.Event, error)
	IsDuplicate(ctx context.Context, deviceID string, eventType domain.EventType, within time.Duration) (bool, error)
}

type CommandRepository interface {
	Store(ctx context.Context, cmd *domain.Command) error
	GetByID(ctx context.Context, id string) (*domain.Command, error)
	ListByDevice(ctx context.Context, deviceID string, limit, offset int) ([]domain.Command, error)
	ListPending(ctx context.Context) ([]domain.Command, error)
	UpdateStatus(ctx context.Context, id string, status domain.CommandStatus, ts time.Time) error
	MarkTimedOut(ctx context.Context, timeout time.Duration) (int64, error)
}

type UserRepository interface {
	Store(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	UpdateRole(ctx context.Context, id int64, role domain.Role) error
	SetActive(ctx context.Context, id int64, active bool) error
}

type DeviceRepository interface {
	Store(ctx context.Context, device *domain.Device) error
	GetByID(ctx context.Context, deviceID string) (*domain.Device, error)
	ListByUser(ctx context.Context, userID int64) ([]domain.Device, error)
	UpdateLastSeen(ctx context.Context, deviceID string, ts time.Time) error
	SetActive(ctx context.Context, deviceID string, active bool) error
}

type TokenRepository interface {
	Store(ctx context.Context, token *domain.Token) error
	GetByDeviceID(ctx context.Context, deviceID string) (*domain.Token, error)
	Revoke(ctx context.Context, id int64) error
	RevokeByDevice(ctx context.Context, deviceID string) error
}
