package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/analytics"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/apikey"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/budget"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/evaluation"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/experiment"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/ingestion"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/localmodel"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/notification"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/portfolio"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/pricing"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/recommendation"
)

type Checker interface{ Ping(context.Context) error }

type Dependencies struct {
	Database, Redis Checker
	Auth            *auth.Service
	Portfolio       *portfolio.Store
	Pricing         *pricing.Store
	APIKeys         *apikey.Store
	Ingestion       *ingestion.Store
	RateLimiter     *redis.Client
	Analytics       *analytics.Store
	Recommendations *recommendation.Store
	Evaluations     *evaluation.Store
	LocalModels     *localmodel.Store
	Experiments     *experiment.Store
	Budgets         *budget.Store
	Notifications   *notification.Store
}

type errorBody struct {
	Error apiError `json:"error"`
}
type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func NewHandler(logger *slog.Logger, deps Dependencies, registry *prometheus.Registry) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	router.Use(requestLog(logger))
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotFound, "not_found", "resource not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})
	router.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/health/ready", readiness(deps))
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	registerAuthRoutes(router, deps.Auth)
	registerPortfolioRoutes(router, deps.Auth, deps.Portfolio)
	registerPricingRoutes(router, deps.Auth, deps.Pricing)
	registerAPIKeyRoutes(router, deps.Auth, deps.APIKeys)
	registerIngestionRoutes(router, deps.Auth, deps.APIKeys, deps.Ingestion, deps.RateLimiter)
	registerAnalyticsRoutes(router, deps.Auth, deps.Analytics)
	registerRecommendationRoutes(router, deps.Auth, deps.Recommendations)
	registerEvaluationRoutes(router, deps.Auth, deps.Evaluations)
	registerLocalModelRoutes(router, deps.Auth, deps.LocalModels)
	registerExperimentRoutes(router, deps.Auth, deps.Experiments)
	registerBudgetRoutes(router, deps.Auth, deps.Budgets)
	registerNotificationRoutes(router, deps.Auth, deps.Notifications)
	return http.MaxBytesHandler(router, 1<<20)
}

func readiness(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if deps.Database == nil || deps.Redis == nil || deps.Database.Ping(ctx) != nil || deps.Redis.Ping(ctx) != nil {
			writeError(w, r, http.StatusServiceUnavailable, "not_ready", "service dependencies are unavailable")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message, RequestID: middleware.GetReqID(r.Context())}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func requestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.InfoContext(r.Context(), "http request", "request_id", middleware.GetReqID(r.Context()), "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
		})
	}
}
