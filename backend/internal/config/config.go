package config

import (
	"fmt"
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
	appEnv := mustGetEnv("APP_ENV")

	return Config{
		AppEnv:         appEnv,
		Port:           mustGetEnv("PORT"),
		FrontendURL:    mustGetEnv("FRONTEND_URL"),
		DatabaseURL:    mustGetEnv("DATABASE_URL"),
		UploadDir:      mustGetEnv("UPLOAD_DIR"),
		MaxUploadBytes: mustGetEnvInt64("MAX_UPLOAD_BYTES"),
		RequestTimeout: time.Duration(mustGetEnvInt64("REQUEST_TIMEOUT_SECONDS")) * time.Second,
		SessionTTL:     time.Duration(mustGetEnvInt64("SESSION_TTL_HOURS")) * time.Hour,
		CookieSecure:   appEnv == "production",
	}
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return value
}

func mustGetEnvInt64(key string) int64 {
	value := mustGetEnv(key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("%s must be an integer", key))
	}

	return parsed
}
