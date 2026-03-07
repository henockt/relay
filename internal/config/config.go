package config

import (
	"os"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SMTPDomain         string // alias address domain, e.g. "relay.example.org"
	SendGridAPIKey     string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "3000"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://"),
		JWTSecret:          getEnv("JWT_SECRET", "secret"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		SMTPDomain:         getEnv("SMTP_DOMAIN", "relay.localhost"),
		SendGridAPIKey:     getEnv("SENDGRID_API_KEY", "dev"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
