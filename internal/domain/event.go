package domain

import "time"

type EventType string

const (
	EventSIMChange        EventType = "sim_change"
	EventBatteryLow       EventType = "battery_low"
	EventWiFiDisconnected EventType = "wifi_disconnected"
	EventPowerOn          EventType = "power_on"
	EventAuthFailure      EventType = "auth_failure"
)

type Event struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	EventType  EventType `json:"event_type"`
	Payload    string    `json:"payload,omitempty"`
	ReceivedAt time.Time `json:"received_at"`
}
