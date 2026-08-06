package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/apikey"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/ingestion"
)

type principalKey struct{}
type batchRequest struct {
	Calls []ingestion.Input `json:"calls"`
}

func registerIngestionRoutes(router chi.Router, authService *auth.Service, keys *apikey.Store, store *ingestion.Store, limiter *redis.Client) {
	if authService == nil || keys == nil || store == nil {
		return
	}
	router.With(apiKeyAuth(keys, limiter, "ingest:calls")).Post("/api/v1/llm-calls", ingestCall(store))
	router.With(apiKeyAuth(keys, limiter, "ingest:calls")).Post("/api/v1/llm-calls/batch", ingestBatch(store))
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/llm-calls", listCalls(store))
		private.Get("/api/v1/llm-calls/{callID}", getCall(store))
	})
}
func apiKeyAuth(keys *apikey.Store, limiter *redis.Client, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, r, 401, "invalid_api_key", "valid API key is required")
				return
			}
			principal, err := keys.Authenticate(r.Context(), strings.TrimPrefix(header, "Bearer "))
			if err != nil || !principal.Scopes[scope] {
				writeError(w, r, 403, "insufficient_scope", "API key does not have the required scope")
				return
			}
			if limiter != nil {
				bucket := time.Now().UTC().Format("200601021504")
				key := "ingest-rate:" + principal.KeyID + ":" + bucket
				count, redisErr := limiter.Incr(r.Context(), key).Result()
				if redisErr != nil {
					writeError(w, r, 503, "rate_limit_unavailable", "ingestion is temporarily unavailable")
					return
				}
				if count == 1 {
					_ = limiter.Expire(r.Context(), key, 2*time.Minute).Err()
				}
				if count > 600 {
					w.Header().Set("Retry-After", "60")
					writeError(w, r, 429, "rate_limit_exceeded", "ingestion rate limit exceeded")
					return
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, principal)))
		})
	}
}
func ingestCall(store *ingestion.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input ingestion.Input
		if json.NewDecoder(r.Body).Decode(&input) != nil || ingestion.Validate(input) != nil {
			writeError(w, r, 400, "invalid_call", "call payload is invalid")
			return
		}
		principal := r.Context().Value(principalKey{}).(apikey.Principal)
		call, err := store.Ingest(r.Context(), principal.OrganizationID, principal.ProjectID, input)
		writeIngestionResult(w, r, call, err, http.StatusCreated)
	}
}
func ingestBatch(store *ingestion.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input batchRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_batch", "batch payload is invalid")
			return
		}
		principal := r.Context().Value(principalKey{}).(apikey.Principal)
		calls, err := store.IngestBatch(r.Context(), principal.OrganizationID, principal.ProjectID, input.Calls)
		writeIngestionResult(w, r, map[string]any{"items": calls, "atomic": true}, err, http.StatusCreated)
	}
}
func listCalls(store *ingestion.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		page, err := store.List(r.Context(), requestClaims(r).OrganizationID, r.URL.Query().Get("project_id"), r.URL.Query().Get("workload_id"), limit, offset)
		writeIngestionResult(w, r, page, err, http.StatusOK)
	}
}
func getCall(store *ingestion.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		call, err := store.Get(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "callID"))
		writeIngestionResult(w, r, call, err, http.StatusOK)
	}
}
func writeIngestionResult(w http.ResponseWriter, r *http.Request, result any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, result)
	case errors.Is(err, ingestion.ErrInvalid):
		writeError(w, r, 400, "invalid_call", "call payload or pricing is invalid")
	case errors.Is(err, ingestion.ErrNotFound):
		writeError(w, r, 422, "dependency_not_found", "project, workload, model, or effective pricing was not found")
	case errors.Is(err, ingestion.ErrConflict):
		writeError(w, r, 409, "idempotency_conflict", "external_call_id was already used with another payload")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
