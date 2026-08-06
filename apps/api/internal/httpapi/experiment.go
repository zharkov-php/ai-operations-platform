package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/experiment"
)

func registerExperimentRoutes(router chi.Router, authService *auth.Service, store *experiment.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/experiments", listExperiments(store))
		private.Get("/api/v1/experiments/{experimentID}", getExperiment(store))
		managed := private.With(requireRoles("owner", "admin", "engineer"))
		managed.Post("/api/v1/experiments", createExperiment(store))
		managed.Post("/api/v1/experiments/{experimentID}/approve", experimentAction(store, "approve"))
		managed.Post("/api/v1/experiments/{experimentID}/start", experimentAction(store, "start"))
		managed.Post("/api/v1/experiments/{experimentID}/pause", experimentAction(store, "pause"))
		managed.Post("/api/v1/experiments/{experimentID}/rollback", experimentAction(store, "rollback"))
		managed.Post("/api/v1/experiments/{experimentID}/complete", completeExperiment(store))
		managed.Post("/api/v1/experiments/{experimentID}/verify", verifyExperiment(store))
		managed.Post("/api/v1/experiments/{experimentID}/metrics", recordExperimentMetrics(store))
	})
}
func listExperiments(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.List(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			writeError(w, r, 500, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	}
}
func getExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, err := store.Get(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "experimentID"))
		writeExperimentResult(w, r, item, err, 200)
	}
}
func createExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input experiment.CreateInput
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "experiment payload is invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.Create(r.Context(), claims.OrganizationID, claims.UserID, input)
		writeExperimentResult(w, r, item, err, 201)
	}
}

type experimentActionRequest struct {
	Reason string `json:"reason"`
}

func experimentAction(store *experiment.Store, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input experimentActionRequest
		_ = json.NewDecoder(r.Body).Decode(&input)
		claims := requestClaims(r)
		item, err := store.Transition(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "experimentID"), action, input.Reason)
		writeExperimentResult(w, r, item, err, 200)
	}
}
func recordExperimentMetrics(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input experiment.Metrics
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "metrics are invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.RecordMetrics(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "experimentID"), input)
		writeExperimentResult(w, r, item, err, 200)
	}
}
func completeExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input experiment.Metrics
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "metrics are invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.RecordMetrics(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "experimentID"), input)
		if err == nil && item.Status == "running" {
			item, err = store.Transition(r.Context(), claims.OrganizationID, claims.UserID, item.ID, "complete", "")
		}
		writeExperimentResult(w, r, item, err, 200)
	}
}
func verifyExperiment(store *experiment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := requestClaims(r)
		item, err := store.Verify(r.Context(), claims.OrganizationID, claims.UserID, chi.URLParam(r, "experimentID"))
		writeExperimentResult(w, r, item, err, 200)
	}
}
func writeExperimentResult(w http.ResponseWriter, r *http.Request, item any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, item)
	case errors.Is(err, experiment.ErrNotFound):
		writeError(w, r, 404, "not_found", "resource not found")
	case errors.Is(err, experiment.ErrInvalidInput):
		writeError(w, r, 400, "invalid_request", "experiment input is invalid")
	case errors.Is(err, experiment.ErrInvalidTransition):
		writeError(w, r, 409, "invalid_transition", "experiment cannot transition")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
