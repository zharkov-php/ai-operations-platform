package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Address      string
	DatabaseURL  string
	RedisURL     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Address:      value("API_ADDRESS", ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     os.Getenv("REDIS_URL"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if cfg.DatabaseURL == "" || cfg.RedisURL == "" {
		return Config{}, errors.New("DATABASE_URL and REDIS_URL are required")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if current := os.Getenv(key); current != "" {
		return current
	}
	return fallback
}
