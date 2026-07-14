package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Est3banj/phone-tracker/internal/domain"
	"github.com/Est3banj/phone-tracker/internal/ports"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	users  ports.UserRepository
	secret string
}

func NewMiddleware(users ports.UserRepository, secret string) *Middleware {
	return &Middleware{
		users:  users,
		secret: secret,
	}
}

// JWTAuth validates JWT tokens and injects user info into context
func (m *Middleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		sub := int64(claims["sub"].(float64))
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		ctx := context.WithValue(r.Context(), ctxKeyUserID, sub)
		ctx = context.WithValue(ctx, ctxKeyUsername, username)
		ctx = context.WithValue(ctx, ctxKeyRole, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LicenseMiddleware checks if the user account is active (licensed)
func (m *Middleware) LicenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(ctxKeyUserID).(int64)
		if !ok || userID == 0 {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		user, err := m.users.GetByID(r.Context(), userID)
		if err != nil || user == nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
			return
		}

		if !user.Active {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "account not activated",
				"message": "Contact admin to activate your account",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AdminOnly restricts to admin/super_admin roles
func (m *Middleware) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ctxKeyRole).(string)
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if role != string(domain.RoleAdmin) && role != string(domain.RoleSuperAdmin) {
			http.Error(w, `{"error":"admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SuperAdminOnly restricts to super_admin role
func (m *Middleware) SuperAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ctxKeyRole).(string)
		if !ok || role != string(domain.RoleSuperAdmin) {
			http.Error(w, `{"error":"super admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORS middleware
func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggerMiddleware logs requests
func (m *Middleware) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple pass-through — use real logger if needed
		next.ServeHTTP(w, r)
	})
}
