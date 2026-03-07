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
	FrontendURL        string
	SendGridAPIKey     string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://"),
		JWTSecret:          getEnv("JWT_SECRET", "secret"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		SMTPDomain:         getEnv("SMTP_DOMAIN", "relay.example.com"),
		SendGridAPIKey:     getEnv("SENDGRID_API_KEY", "dev"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
