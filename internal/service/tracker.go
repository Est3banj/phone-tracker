package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
	"github.com/Est3banj/phone-tracker/internal/ports"
)

type Tracker struct {
	locations ports.LocationRepository
	events    ports.EventRepository
	notifier  ports.Notifier
}

func NewTracker(locRepo ports.LocationRepository, evtRepo ports.EventRepository, notifier ports.Notifier) *Tracker {
	return &Tracker{
		locations: locRepo,
		events:    evtRepo,
		notifier:  notifier,
	}
}

func (t *Tracker) ProcessLocation(ctx context.Context, loc *domain.LocationReport) error {
	loc.ReceivedAt = time.Now().UTC()

	if err := t.locations.Store(ctx, loc); err != nil {
		return err
	}

	msg, err := json.Marshal(map[string]interface{}{
		"type":     "location",
		"ts":       loc.ReceivedAt.Format(time.RFC3339),
		"lat":      loc.Latitude,
		"lng":      loc.Longitude,
		"alt":      loc.Altitude,
		"accuracy": loc.Accuracy,
		"speed":    loc.Speed,
		"battery":  loc.Battery,
		"charging": loc.Charging,
	})
	if err != nil {
		return err
	}

	log.Printf("Location from %s: (%.6f, %.6f) bat=%d%%", loc.DeviceID, loc.Latitude, loc.Longitude, loc.Battery)

	// Notify dashboards
	t.notifier.Broadcast(msg)
	return nil
}

func (t *Tracker) ProcessEvent(ctx context.Context, evt *domain.Event) error {
	// Deduplicate same-type events within 5 minutes
	dup, err := t.events.IsDuplicate(ctx, evt.DeviceID, evt.EventType, 5*time.Minute)
	if err != nil {
		return err
	}
	if dup {
		log.Printf("Dropping duplicate event %s for device %s", evt.EventType, evt.DeviceID)
		return nil
	}

	evt.ReceivedAt = time.Now().UTC()
	if err := t.events.Store(ctx, evt); err != nil {
		return err
	}

	msg, err := json.Marshal(map[string]interface{}{
		"type":       "event",
		"ts":         evt.ReceivedAt.Format(time.RFC3339),
		"event_type": evt.EventType,
		"payload":    evt.Payload,
	})
	if err != nil {
		return err
	}

	t.notifier.Broadcast(msg)
	return nil
}
