package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/config"
)

func TestRunShutsDownWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg := config.Config{
		Address: "127.0.0.1:0", DatabaseURL: "postgres://localhost/example", RedisURL: "redis://localhost:6379/0",
		ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
	}
	if err := Run(ctx, cfg, slog.New(slog.NewJSONHandler(io.Discard, nil))); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
