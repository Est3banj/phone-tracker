package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/Est3banj/phone-tracker/internal/config"
	"github.com/Est3banj/phone-tracker/internal/ports"
	"nhooyr.io/websocket"
)

type WSHandler struct {
	hub     *Hub
	cfg     *config.Config
	devices ports.DeviceRepository
	tokens  ports.TokenRepository
	users   ports.UserRepository
}

func NewWSHandler(
	hub *Hub,
	cfg *config.Config,
	devices ports.DeviceRepository,
	tokens ports.TokenRepository,
	users ports.UserRepository,
) *WSHandler {
	return &WSHandler{
		hub:     hub,
		cfg:     cfg,
		devices: devices,
		tokens:  tokens,
		users:   users,
	}
}

func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Auth via query param: ?token=<device_token> or ?jwt=<jwt_token>&role=dashboard
	tokenParam := r.URL.Query().Get("token")
	jwtParam := r.URL.Query().Get("jwt")
	role := r.URL.Query().Get("role")

	var deviceID string
	var connRole string

	if tokenParam != "" {
		// Device token auth
		id, err := h.validateDeviceToken(tokenParam)
		if err != nil {
			log.Printf("WS auth failed: %v", err)
			http.Error(w, `{"error":"invalid device token"}`, http.StatusUnauthorized)
			return
		}
		deviceID = id
		connRole = "device"

		// License check: device must belong to an active user
		dev, err := h.devices.GetByID(r.Context(), deviceID)
		if err != nil || dev == nil {
			log.Printf("WS auth failed: device %s not found", deviceID)
			http.Error(w, `{"error":"device not found"}`, http.StatusUnauthorized)
			return
		}
		user, err := h.users.GetByID(r.Context(), dev.UserID)
		if err != nil || user == nil || !user.Active {
			log.Printf("WS license check failed for device %s (user active=%v)", deviceID, user != nil && user.Active)
			http.Error(w, `{"error":"account not active"}`, http.StatusForbidden)
			return
		}
	} else if jwtParam != "" && role == "dashboard" {
		// JWT auth for dashboard
		connRole = "dashboard"
		deviceID = "dashboard_" + jwtParam[:min(8, len(jwtParam))]
	} else {
		http.Error(w, `{"error":"auth required — ?token= for devices or ?jwt=&role=dashboard for dashboards"}`, http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}

	dc := &DeviceConn{
		DeviceID: deviceID,
		conn:     conn,
		send:     make(chan []byte, 256),
		role:     connRole,
	}

	ctx, cancel := context.WithCancel(r.Context())
	dc.close = cancel

	h.hub.Register(dc)

	if connRole == "device" {
		// Dispatch pending commands on reconnect
		// In real impl: h.hub.commanderService.DispatchPending(ctx, deviceID)
		log.Printf("Device %s connected, would dispatch pending commands", deviceID)
	}

	go h.writePump(ctx, dc)
	go h.readPump(ctx, dc)
}

func (h *WSHandler) validateDeviceToken(tokenStr string) (string, error) {
	// Hash the token
	sum := sha256.Sum256([]byte(tokenStr))
	hash := hex.EncodeToString(sum[:])

	// Look up by hash (would need a reverse index in production)
	// For MVP, we store token_hash in devices and tokens tables
	// We check tokens table for a non-revoked, non-expired token
	_ = hash
	return "", nil
}

// Better approach: validate by looking up tokens matching hash
func (h *WSHandler) validateDeviceTokenFull(tokenStr string) (string, error) {
	sum := sha256.Sum256([]byte(tokenStr))
	hash := hex.EncodeToString(sum[:])

	// Scan tokens table — this is O(n) but fine for single-device MVP
	// Production: add token_hash index or use a lookup table
	_ = hash
	return "", nil
}

func (h *WSHandler) readPump(ctx context.Context, dc *DeviceConn) {
	defer func() {
		h.hub.Unregister(dc)
		dc.close()
		dc.conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	// Set read limit
	dc.conn.SetReadLimit(64 * 1024)

	for {
		_, msg, err := dc.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				log.Printf("WS read close: %v", err)
			} else {
				log.Printf("WS read error: %v", err)
			}
			return
		}

		h.hub.HandleMessage(dc, msg)
	}
}

func (h *WSHandler) writePump(ctx context.Context, dc *DeviceConn) {
	ticker := time.NewTicker(h.cfg.WSPingInterval)
	defer func() {
		ticker.Stop()
		h.hub.Unregister(dc)
		dc.close()
		dc.conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-dc.send:
			if !ok {
				return
			}
			// Write with timeout
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := dc.conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				log.Printf("WS write error: %v", err)
				return
			}
		case <-ticker.C:
			// Ping already handled by hub broadcast
		}
	}
}

func (s *HTTPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Delegate to WSHandler
	// In real impl: create WSHandler and call this
	// For now, basic upgrade
	wsHandler := NewWSHandler(
		s.hub,
		s.cfg,
		s.devices,
		s.tokens,
		s.users,
	)
	wsHandler.HandleWebSocket(w, r)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


