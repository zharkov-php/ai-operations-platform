package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Address         string
	DatabaseURL     string
	RedisURL        string
	AuthTokenSecret string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Address:         value("API_ADDRESS", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		AuthTokenSecret: os.Getenv("AUTH_TOKEN_SECRET"),
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
	if cfg.DatabaseURL == "" || cfg.RedisURL == "" || len(cfg.AuthTokenSecret) < 32 {
		return Config{}, errors.New("DATABASE_URL, REDIS_URL, and an AUTH_TOKEN_SECRET of at least 32 characters are required")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if current := os.Getenv(key); current != "" {
		return current
	}
	return fallback
}
