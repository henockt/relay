package config

import (
	"os"
)
type Config struct {
	Port string
	DatabaseURL string
	JWTSecret string
	GoogleClientID string
	GoogleClientSecret string
	GoogleRedirectURL string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "3000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://"),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL: getEnv("GOOGLE_REDIRECT_URL", ""),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}