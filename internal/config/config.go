package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	ClerkSecretKey    string
	SMTPDomain        string // alias address domain, e.g. "relay.example.org"
	FrontendURL       string
	MailgunAPIKey     string
	MailgunDomain     string // Mailgun sending domain, usually the same as SMTPDomain
	MailgunAPIBase    string // https://api.mailgun.net/v3, or the EU endpoint
	MailgunSigningKey string // verifies inbound webhook signatures
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://"),
		ClerkSecretKey:    getEnv("CLERK_SECRET_KEY", ""),
		SMTPDomain:        getEnv("SMTP_DOMAIN", "relay.example.com"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		MailgunAPIKey:     getEnv("MAILGUN_API_KEY", "dev"),
		MailgunDomain:     getEnv("MAILGUN_DOMAIN", getEnv("SMTP_DOMAIN", "relay.example.com")),
		MailgunAPIBase:    getEnv("MAILGUN_API_BASE", "https://api.mailgun.net/v3"),
		MailgunSigningKey: getEnv("MAILGUN_SIGNING_KEY", ""),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
