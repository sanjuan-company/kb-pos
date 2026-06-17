package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port    string
	LogLevel string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string

	RetryMaxAttempts    int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
		FromName:     getEnv("FROM_NAME", "No Reply"),

		RetryMaxAttempts:    getEnvInt("RETRY_MAX_ATTEMPTS", 3),
		RetryInitialBackoff: getEnvDuration("RETRY_INITIAL_BACKOFF", 1*time.Second),
		RetryMaxBackoff:     getEnvDuration("RETRY_MAX_BACKOFF", 30*time.Second),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return v
}
