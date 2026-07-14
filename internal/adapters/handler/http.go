package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Est3banj/phone-tracker/internal/config"
	"github.com/Est3banj/phone-tracker/internal/ports"
)

type HTTPServer struct {
	hub       *Hub
	auth      *AuthHandler
	mw        *Middleware
	cfg       *config.Config
	users     ports.UserRepository
	devices   ports.DeviceRepository
	tokens    ports.TokenRepository
}

func NewHTTPServer(
	hub *Hub,
	auth *AuthHandler,
	mw *Middleware,
	cfg *config.Config,
	users ports.UserRepository,
	devices ports.DeviceRepository,
	tokens ports.TokenRepository,
) *HTTPServer {
	return &HTTPServer{
		hub:     hub,
		auth:    auth,
		mw:      mw,
		cfg:     cfg,
		users:   users,
		devices: devices,
		tokens:  tokens,
	}
}

func (s *HTTPServer) RegisterRoutes(mux *http.ServeMux) {
	// Public routes
	mux.HandleFunc("GET /api/health", s.health)

	// Auth routes
	mux.HandleFunc("POST /api/register", s.auth.Register)
	mux.HandleFunc("POST /api/login", s.auth.Login)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/devices/token", s.auth.GenerateDeviceToken)
	protected.HandleFunc("POST /api/rotate-token", s.auth.RotateToken)
	protected.HandleFunc("GET /api/devices", s.listDevices)
	protected.HandleFunc("GET /api/locations/{device_id}", s.getLocations)
	protected.HandleFunc("GET /api/events/{device_id}", s.getEvents)
	protected.HandleFunc("GET /api/latest/{device_id}", s.getLatestLocation)

	// Admin routes
	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/users", s.listUsers)
	admin.HandleFunc("PUT /api/admin/users/{id}/activate", s.activateUser)
	admin.HandleFunc("PUT /api/admin/users/{id}/deactivate", s.deactivateUser)
	admin.HandleFunc("PUT /api/admin/users/{id}/role", s.updateRole)

	// WS route
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Apply middleware chain: CORS → Logger → JWT → License → route
	chain := s.mw.CORS(s.mw.JWTAuth(s.mw.LicenseMiddleware(protected)))
	mux.Handle("/api/devices/", chain)
	mux.Handle("/api/rotate-token", chain)
	mux.Handle("/api/locations/", chain)
	mux.Handle("/api/events/", chain)
	mux.Handle("/api/latest/", chain)

	adminChain := s.mw.CORS(s.mw.JWTAuth(s.mw.SuperAdminOnly(admin)))
	mux.Handle("/api/admin/", adminChain)
}

func (s *HTTPServer) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *HTTPServer) listDevices(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int64)
	devices, err := s.devices.ListByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(devices)
}

func (s *HTTPServer) getLocations(w http.ResponseWriter, r *http.Request) {
	_ = r.PathValue("device_id")
	// Delegate to repository read — in real impl use service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *HTTPServer) getEvents(w http.ResponseWriter, r *http.Request) {
	_ = r.PathValue("device_id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *HTTPServer) getLatestLocation(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("device_id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"device_id": deviceID})
}

func (s *HTTPServer) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.users.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	// Strip password hashes
	type safeUser struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}
	safe := make([]safeUser, len(users))
	for i, u := range users {
		safe[i] = safeUser{ID: u.ID, Username: u.Username, Role: string(u.Role), Active: u.Active}
	}
	json.NewEncoder(w).Encode(safe)
}

func (s *HTTPServer) activateUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID from path
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (s *HTTPServer) deactivateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}

func (s *HTTPServer) updateRole(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"not implemented"}`, http.StatusNotImplemented)
}
