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

	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/analytics"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/apikey"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/budget"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/config"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/dependencies"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/evaluation"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/experiment"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/httpapi"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/ingestion"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/localmodel"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/portfolio"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/pricing"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/recommendation"
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
	pricingStore := pricing.NewStore(pool)
	server := &http.Server{Addr: cfg.Address, Handler: httpapi.NewHandler(logger, httpapi.Dependencies{Database: dependencies.Postgres{Pool: pool}, Redis: dependencies.Redis{Client: redisClient}, Auth: authService, Portfolio: portfolio.NewStore(pool), Pricing: pricingStore, APIKeys: apikey.NewStore(pool), Ingestion: ingestion.NewStore(pool, pricingStore), RateLimiter: redisClient, Analytics: analytics.NewStore(pool), Recommendations: recommendation.NewStore(pool), Evaluations: evaluation.NewStore(pool), LocalModels: localmodel.NewStore(pool), Experiments: experiment.NewStore(pool), Budgets: budget.NewStore(pool)}, prometheus.NewRegistry()), ReadHeaderTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout}
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
