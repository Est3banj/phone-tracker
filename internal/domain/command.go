package domain

import "time"

type CommandAction string

const (
	CmdLockDevice  CommandAction = "lock_device"
	CmdWipeDevice  CommandAction = "wipe_device"
	CmdCapturePhoto CommandAction = "capture_photo"
	CmdTriggerAlarm CommandAction = "trigger_alarm"
)

type CommandStatus string

const (
	CmdPending      CommandStatus = "pending"
	CmdSent         CommandStatus = "sent"
	CmdReceived     CommandStatus = "received"
	CmdExecuted     CommandStatus = "executed"
	CmdFailed       CommandStatus = "failed"
	CmdTimedOut     CommandStatus = "timed_out"
)

type Command struct {
	ID              string        `json:"id"`
	DeviceID        string        `json:"device_id"`
	Action          CommandAction `json:"action"`
	Params          string        `json:"params,omitempty"`
	Status          CommandStatus `json:"status"`
	CreatedAt       time.Time     `json:"created_at"`
	SentAt          *time.Time    `json:"sent_at,omitempty"`
	AcknowledgedAt  *time.Time    `json:"acknowledged_at,omitempty"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	Error           *string       `json:"error,omitempty"`
}
