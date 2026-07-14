package handler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Est3banj/phone-tracker/internal/domain"
	"github.com/Est3banj/phone-tracker/internal/ports"
	"github.com/Est3banj/phone-tracker/internal/service"
	"nhooyr.io/websocket"
)

type DeviceConn struct {
	DeviceID string
	conn     *websocket.Conn
	send     chan []byte
	close    func()
	role     string // "device" or "dashboard"
}

type Hub struct {
	mu      sync.RWMutex
	devices map[string]*DeviceConn
	dashboards []*DeviceConn

	trackerService   *service.Tracker
	commanderService *service.Commander
	locationRepo     ports.LocationRepository
	eventRepo        ports.EventRepository
}

func NewHub() *Hub {
	return &Hub{
		devices:   make(map[string]*DeviceConn),
		dashboards: make([]*DeviceConn, 0),
	}
}

func (h *Hub) SetTracker(svc *service.Tracker) {
	h.trackerService = svc
}

func (h *Hub) SetCommander(svc *service.Commander) {
	h.commanderService = svc
}

func (h *Hub) SetRepos(locRepo ports.LocationRepository, evtRepo ports.EventRepository) {
	h.locationRepo = locRepo
	h.eventRepo = evtRepo
}

func (h *Hub) Register(dc *DeviceConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if dc.role == "device" {
		if existing, ok := h.devices[dc.DeviceID]; ok {
			log.Printf("Replacing existing connection for device %s", dc.DeviceID)
			existing.close()
		}
		h.devices[dc.DeviceID] = dc
		log.Printf("Device %s registered", dc.DeviceID)
	} else {
		h.dashboards = append(h.dashboards, dc)
		log.Printf("Dashboard registered")
	}
}

func (h *Hub) Unregister(dc *DeviceConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if dc.role == "device" {
		if existing, ok := h.devices[dc.DeviceID]; ok && existing == dc {
			delete(h.devices, dc.DeviceID)
			log.Printf("Device %s unregistered", dc.DeviceID)
		}
	} else {
		for i, d := range h.dashboards {
			if d == dc {
				h.dashboards = append(h.dashboards[:i], h.dashboards[i+1:]...)
				break
			}
		}
		log.Printf("Dashboard unregistered")
	}
}

func (h *Hub) SendToDevice(deviceID string, msg []byte) error {
	h.mu.RLock()
	dc, ok := h.devices[deviceID]
	h.mu.RUnlock()

	if !ok {
		log.Printf("Device %s not connected", deviceID)
		return nil
	}

	select {
	case dc.send <- msg:
		return nil
	default:
		log.Printf("Send buffer full for device %s, dropping message", deviceID)
		return nil
	}
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, dc := range h.dashboards {
		select {
		case dc.send <- msg:
		default:
			log.Printf("Dashboard send buffer full, dropping message")
		}
	}
}

func (h *Hub) IsConnected(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.devices[deviceID]
	return ok
}

func (h *Hub) HandleMessage(dc *DeviceConn, msg []byte) {
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		log.Printf("Invalid message from %s: %v", dc.DeviceID, err)
		return
	}

	switch envelope.Type {
	case "location":
		if dc.role != "device" {
			return
		}
		var loc domain.LocationReport
		if err := json.Unmarshal(msg, &loc); err != nil {
			log.Printf("Invalid location report: %v", err)
			return
		}
		loc.DeviceID = dc.DeviceID
		if h.trackerService != nil {
			ctx := &simpleContext{}
			if err := h.trackerService.ProcessLocation(ctx, &loc); err != nil {
				log.Printf("Error processing location: %v", err)
			}
		}

	case "event":
		if dc.role != "device" {
			return
		}
		var evt struct {
			EventType string `json:"event_type"`
			Payload   string `json:"payload"`
		}
		if err := json.Unmarshal(msg, &evt); err != nil {
			log.Printf("Invalid event: %v", err)
			return
		}
		domainEvt := &domain.Event{
			DeviceID:  dc.DeviceID,
			EventType: domain.EventType(evt.EventType),
			Payload:   evt.Payload,
		}
		if h.trackerService != nil {
			ctx := &simpleContext{}
			if err := h.trackerService.ProcessEvent(ctx, domainEvt); err != nil {
				log.Printf("Error processing event: %v", err)
			}
		}

	case "pong":
		// Keepalive response — nothing to do

	case "ack":
		if dc.role != "device" {
			return
		}
		var ack struct {
			CmdID string `json:"cmd_id"`
		}
		if err := json.Unmarshal(msg, &ack); err != nil {
			log.Printf("Invalid ack: %v", err)
			return
		}
		if h.commanderService != nil {
			ctx := &simpleContext{}
			if err := h.commanderService.HandleAck(ctx, ack.CmdID); err != nil {
				log.Printf("Error handling ack: %v", err)
			}
		}

	case "result":
		if dc.role != "device" {
			return
		}
		var result struct {
			CmdID  string  `json:"cmd_id"`
			Status string  `json:"status"`
			Error  *string `json:"error"`
		}
		if err := json.Unmarshal(msg, &result); err != nil {
			log.Printf("Invalid result: %v", err)
			return
		}
		if h.commanderService != nil {
			ctx := &simpleContext{}
			status := domain.CommandStatus(result.Status)
			if err := h.commanderService.HandleResult(ctx, result.CmdID, status, result.Error); err != nil {
				log.Printf("Error handling result: %v", err)
			}
		}
	}
}

// Hub ticker for sending pings
func (h *Hub) PingLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ping, _ := json.Marshal(map[string]string{"type": "ping"})
			h.Broadcast(ping)

			h.mu.RLock()
			for _, dc := range h.devices {
				select {
				case dc.send <- ping:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

type simpleContext struct{}

func (s *simpleContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (s *simpleContext) Done() <-chan struct{}       { return nil }
func (s *simpleContext) Err() error                  { return nil }
func (s *simpleContext) Value(interface{}) interface{} { return nil }
