package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/recommendation"
)

func registerRecommendationRoutes(router chi.Router, authService *auth.Service, store *recommendation.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/recommendations", listRecommendations(store))
		private.Get("/api/v1/recommendations/{recommendationID}", getRecommendation(store))
		private.Get("/api/v1/recommendations/{recommendationID}/audit-history", recommendationAuditHistory(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/recommendations/analyze", analyzeRecommendations(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/recommendations/{recommendationID}/accept", reviewRecommendation(store, "accept"))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/recommendations/{recommendationID}/reject", reviewRecommendation(store, "reject"))
	})
}

type recommendationReviewRequest struct {
	Reason string `json:"reason"`
}

func reviewRecommendation(store *recommendation.Store, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input recommendationReviewRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
				writeError(w, r, http.StatusBadRequest, "invalid_request", "review payload is invalid")
				return
			}
		}
		claims := requestClaims(r)
		item, err := store.Review(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "recommendationID"), action, input.Reason)
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, item)
		case errors.Is(err, recommendation.ErrNotFound):
			writeError(w, r, http.StatusNotFound, "not_found", "resource not found")
		case errors.Is(err, recommendation.ErrReasonRequired):
			writeError(w, r, http.StatusBadRequest, "reason_required", "a rejection reason between 1 and 1000 characters is required")
		case errors.Is(err, recommendation.ErrInvalidTransition):
			writeError(w, r, http.StatusConflict, "invalid_transition", "recommendation has already been reviewed or cannot transition")
		default:
			writeError(w, r, http.StatusInternalServerError, "internal_error", "request could not be completed")
		}
	}
}

func recommendationAuditHistory(store *recommendation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.AuditHistory(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "recommendationID"))
		if errors.Is(err, recommendation.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}
func analyzeRecommendations(store *recommendation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.AnalyzeOrganization(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			writeError(w, r, 500, "analysis_failed", "recommendation analysis could not be completed")
			return
		}
		writeJSON(w, 200, map[string]any{"items": items, "rule_version": recommendation.RuleVersion, "savings_status": "estimated"})
	}
}
func listRecommendations(store *recommendation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.List(r.Context(), requestClaims(r).OrganizationID, r.URL.Query().Get("status"), r.URL.Query().Get("priority"))
		if err != nil {
			writeError(w, r, 500, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	}
}
func getRecommendation(store *recommendation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, err := store.Get(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "recommendationID"))
		if errors.Is(err, recommendation.ErrNotFound) {
			writeError(w, r, 404, "not_found", "resource not found")
			return
		}
		if err != nil {
			writeError(w, r, 500, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, 200, item)
	}
}
