package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv         string
	Port           string
	FrontendURL    string
	DatabaseURL    string
	UploadDir      string
	MaxUploadBytes int64
	RequestTimeout time.Duration
	SessionTTL     time.Duration
	CookieSecure   bool
}

func Load() Config {
	return Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://knowledge:knowledge@localhost:5432/knowledge_assistant?sslmode=disable"),
		UploadDir:      getEnv("UPLOAD_DIR", "./data/uploads"),
		MaxUploadBytes: getEnvInt64("MAX_UPLOAD_BYTES", 10*1024*1024),
		RequestTimeout: time.Duration(getEnvInt64("REQUEST_TIMEOUT_SECONDS", 30)) * time.Second,
		SessionTTL:     time.Duration(getEnvInt64("SESSION_TTL_HOURS", 24)) * time.Hour,
		CookieSecure:   getEnv("APP_ENV", "development") == "production",
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
