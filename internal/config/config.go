package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddr    string
	DatabasePath  string
	JWTSecret     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	AdminUsername  string
	AdminPassword  string
	WSPingInterval time.Duration
	WSTimeout     time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddr:    getEnv("SERVER_ADDR", ":8080"),
		DatabasePath:  getEnv("DB_PATH", "phone-tracker.db"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production-32bytes!"),
		TokenExpiry:   getDurationEnv("TOKEN_EXPIRY", 24*time.Hour),
		RefreshExpiry: getDurationEnv("REFRESH_EXPIRY", 30*24*time.Hour),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin"),
		WSPingInterval: getDurationEnv("WS_PING_INTERVAL", 30*time.Second),
		WSTimeout:     getDurationEnv("WS_TIMEOUT", 90*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
