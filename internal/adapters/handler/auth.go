package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Est3banj/phone-tracker/internal/config"
	"github.com/Est3banj/phone-tracker/internal/domain"
	"github.com/Est3banj/phone-tracker/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg      *config.Config
	users    ports.UserRepository
	devices  ports.DeviceRepository
	tokens   ports.TokenRepository
}

func NewAuthHandler(cfg *config.Config, users ports.UserRepository, devices ports.DeviceRepository, tokens ports.TokenRepository) *AuthHandler {
	return &AuthHandler{
		cfg:     cfg,
		users:   users,
		devices: devices,
		tokens:  tokens,
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	DeviceToken  string `json:"device_token,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		http.Error(w, `{"error":"username min 3 chars, password min 6 chars"}`, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	user := &domain.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		Active:       false, // Must be activated by super_admin
	}

	if err := h.users.Store(r.Context(), user); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, `{"error":"username already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user created, awaiting activation"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.users.GetByUsername(r.Context(), req.Username)
	if err != nil || user == nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if !user.Active {
		http.Error(w, `{"error":"account not activated"}`, http.StatusForbidden)
		return
	}

	token, err := h.generateJWT(user)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.generateRefreshToken()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(authResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) RotateToken(w http.ResponseWriter, r *http.Request) {
	deviceID := r.Context().Value(ctxKeyDeviceID).(string)
	if deviceID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Revoke old tokens
	if err := h.tokens.RevokeByDevice(r.Context(), deviceID); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// Generate new token
	tokenStr, hash, err := generateDeviceToken()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	deviceToken := &domain.Token{
		DeviceID:  deviceID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
	}

	if err := h.tokens.Store(r.Context(), deviceToken); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(authResponse{
		DeviceToken: tokenStr,
	})
}

func (h *AuthHandler) GenerateDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID).(int64)
	if userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	deviceID := fmt.Sprintf("dev_%x_%d", time.Now().UnixNano(), userID)

	tokenStr, hash, err := generateDeviceToken()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	dev := &domain.Device{
		DeviceID:  deviceID,
		UserID:    userID,
		TokenHash: hash,
		Active:    true,
	}

	if err := h.devices.Store(r.Context(), dev); err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	deviceToken := &domain.Token{
		DeviceID:  deviceID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
	}
	h.tokens.Store(r.Context(), deviceToken)

	json.NewEncoder(w).Encode(authResponse{
		DeviceToken: tokenStr,
	})
}

func (h *AuthHandler) generateJWT(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     string(user.Role),
		"exp":      time.Now().Add(h.cfg.TokenExpiry).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *AuthHandler) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateDeviceToken() (string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := "pt_v1_" + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

// JWT validation for request context
type contextKey string

const (
	ctxKeyUserID   contextKey = "user_id"
	ctxKeyUsername  contextKey = "username"
	ctxKeyRole     contextKey = "role"
	ctxKeyDeviceID contextKey = "device_id"
)
