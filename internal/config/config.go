package config

import (
	"os"
)

type Config struct {
	Port string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "3000"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}