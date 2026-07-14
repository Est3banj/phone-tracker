package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
	"github.com/Est3banj/phone-tracker/internal/ports"
)

type Commander struct {
	commands ports.CommandRepository
	notifier ports.Notifier
}

func NewCommander(cmdRepo ports.CommandRepository, notifier ports.Notifier) *Commander {
	return &Commander{
		commands: cmdRepo,
		notifier: notifier,
	}
}

func (c *Commander) Issue(ctx context.Context, deviceID string, action domain.CommandAction, params string) (*domain.Command, error) {
	id, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	cmd := &domain.Command{
		ID:        id,
		DeviceID:  deviceID,
		Action:    action,
		Params:    params,
		Status:    domain.CmdPending,
		CreatedAt: time.Now().UTC(),
	}

	if err := c.commands.Store(ctx, cmd); err != nil {
		return nil, fmt.Errorf("store command: %w", err)
	}

	if c.notifier.IsConnected(deviceID) {
		_ = c.dispatch(ctx, cmd)
	} else {
		log.Printf("Command %s stored as pending (device %s offline)", cmd.ID, deviceID)
	}

	return cmd, nil
}

func (c *Commander) dispatch(ctx context.Context, cmd *domain.Command) error {
	msg, err := json.Marshal(map[string]interface{}{
		"type":    "command",
		"cmd_id":  cmd.ID,
		"action":  cmd.Action,
		"params":  cmd.Params,
		"ts":      cmd.CreatedAt.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	if err := c.notifier.SendToDevice(cmd.DeviceID, msg); err != nil {
		return err
	}

	now := time.Now().UTC()
	return c.commands.UpdateStatus(ctx, cmd.ID, domain.CmdSent, now)
}

func (c *Commander) HandleAck(ctx context.Context, cmdID string) error {
	now := time.Now().UTC()
	return c.commands.UpdateStatus(ctx, cmdID, domain.CmdReceived, now)
}

func (c *Commander) HandleResult(ctx context.Context, cmdID string, status domain.CommandStatus, errMsg *string) error {
	now := time.Now().UTC()
	if err := c.commands.UpdateStatus(ctx, cmdID, status, now); err != nil {
		return err
	}

	cmd, err := c.commands.GetByID(ctx, cmdID)
	if err != nil || cmd == nil {
		return err
	}

	msg, err := json.Marshal(map[string]interface{}{
		"type":   "cmd_status",
		"cmd_id": cmdID,
		"status": status,
		"error":  errMsg,
	})
	if err != nil {
		return err
	}
	c.notifier.Broadcast(msg)
	return nil
}

func (c *Commander) DispatchPending(ctx context.Context, deviceID string) {
	cmds, err := c.commands.ListPending(ctx)
	if err != nil {
		log.Printf("Error listing pending commands: %v", err)
		return
	}
	for _, cmd := range cmds {
		if cmd.DeviceID == deviceID {
			if err := c.dispatch(ctx, &cmd); err != nil {
				log.Printf("Error dispatching pending command %s: %v", cmd.ID, err)
			}
		}
	}
}

func (c *Commander) TimeoutLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := c.commands.MarkTimedOut(ctx, 60*time.Second)
			if err != nil {
				log.Printf("Error marking timed out commands: %v", err)
			} else if count > 0 {
				log.Printf("Timed out %d commands", count)
			}
		}
	}
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
