package httpapi

import (
	"errors"
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
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/recommendations/analyze", analyzeRecommendations(store))
	})
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
