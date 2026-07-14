package domain

import "time"

type LocationReport struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	Latitude   float64   `json:"lat"`
	Longitude  float64   `json:"lng"`
	Altitude   *float64  `json:"alt,omitempty"`
	Accuracy   *float64  `json:"accuracy,omitempty"`
	Speed      *float64  `json:"speed,omitempty"`
	Battery    int       `json:"battery"`
	Charging   bool      `json:"charging"`
	ReceivedAt time.Time `json:"received_at"`
}

type LocationFilter struct {
	DeviceID string
	From     time.Time
	To       time.Time
	Limit    int
	Offset   int
}
