package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/config"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/dependencies"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/httpapi"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/portfolio"
)

func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	authService := auth.NewService(auth.NewPostgresStore(pool), []byte(cfg.AuthTokenSecret))
	server := &http.Server{Addr: cfg.Address, Handler: httpapi.NewHandler(logger, httpapi.Dependencies{Database: dependencies.Postgres{Pool: pool}, Redis: dependencies.Redis{Client: redisClient}, Auth: authService, Portfolio: portfolio.NewStore(pool)}, prometheus.NewRegistry()), ReadHeaderTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	logger.Info("api started", "address", cfg.Address)
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
