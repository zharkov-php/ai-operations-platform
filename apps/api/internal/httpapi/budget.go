package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/budget"
	"net/http"
)

func registerBudgetRoutes(router chi.Router, authService *auth.Service, store *budget.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/budget-thresholds", func(w http.ResponseWriter, r *http.Request) {
			items, err := store.ListThresholds(r.Context(), requestClaims(r).OrganizationID)
			if err != nil {
				writeError(w, r, 500, "internal_error", "request could not be completed")
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		private.Get("/api/v1/budget-alerts", func(w http.ResponseWriter, r *http.Request) {
			items, err := store.ListAlerts(r.Context(), requestClaims(r).OrganizationID, r.URL.Query().Get("status"))
			if err != nil {
				writeError(w, r, 500, "internal_error", "request could not be completed")
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		private.With(requireRoles("owner", "admin", "finance")).Post("/api/v1/projects/{projectID}/budget-thresholds", setBudgetThreshold(store))
		private.With(requireRoles("owner", "admin", "engineer", "finance")).Post("/api/v1/budget-alerts/{alertID}/acknowledge", acknowledgeBudgetAlert(store))
		private.With(requireRoles("owner", "admin", "finance")).Post("/api/v1/budget-alerts/evaluate", evaluateBudgets(store))
	})
}

type thresholdRequest struct {
	Percentage string `json:"percentage"`
	Severity   string `json:"severity"`
}

func setBudgetThreshold(store *budget.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input thresholdRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "threshold is invalid")
			return
		}
		item, err := store.SetThreshold(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "projectID"), input.Percentage, input.Severity)
		writeBudgetResult(w, r, item, err, 201)
	}
}
func acknowledgeBudgetAlert(store *budget.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := requestClaims(r)
		item, err := store.Acknowledge(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "alertID"))
		writeBudgetResult(w, r, item, err, 200)
	}
}
func evaluateBudgets(store *budget.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		created, suppressed, err := store.EvaluateOrganization(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			writeError(w, r, 500, "internal_error", "budget evaluation failed")
			return
		}
		writeJSON(w, 200, map[string]any{"created": created, "suppressed": suppressed})
	}
}
func writeBudgetResult(w http.ResponseWriter, r *http.Request, item any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, item)
	case errors.Is(err, budget.ErrInvalidInput):
		writeError(w, r, 400, "invalid_request", "budget input is invalid")
	case errors.Is(err, budget.ErrNotFound):
		writeError(w, r, 404, "not_found", "resource not found")
	case errors.Is(err, budget.ErrInvalidTransition):
		writeError(w, r, 409, "invalid_transition", "alert is not open")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
