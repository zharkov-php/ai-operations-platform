package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/evaluation"
)

func registerEvaluationRoutes(router chi.Router, authService *auth.Service, store *evaluation.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/evaluation-datasets", listEvaluationDatasets(store))
		private.Get("/api/v1/evaluation-datasets/{datasetID}", getEvaluationDataset(store))
		private.Get("/api/v1/evaluation-runs/{runID}", getEvaluationRun(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/evaluation-datasets", createEvaluationDataset(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/evaluation-runs", createEvaluationRun(store))
	})
}

func createEvaluationDataset(store *evaluation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input evaluation.CreateDatasetInput
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, http.StatusBadRequest, "invalid_request", "evaluation dataset payload is invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.CreateDataset(r.Context(), claims.OrganizationID, claims.UserID, input)
		writeEvaluationResult(w, r, item, err, http.StatusCreated)
	}
}

func listEvaluationDatasets(store *evaluation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.ListDatasets(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func getEvaluationDataset(store *evaluation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, err := store.GetDataset(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "datasetID"))
		writeEvaluationResult(w, r, item, err, http.StatusOK)
	}
}

type createRunRequest struct {
	DatasetID          string         `json:"dataset_id"`
	CandidateExecution map[string]any `json:"candidate_execution"`
}

func createEvaluationRun(store *evaluation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createRunRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.DatasetID == "" {
			writeError(w, r, http.StatusBadRequest, "invalid_request", "evaluation run payload is invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.CreateRun(r.Context(), claims.OrganizationID, claims.UserID, input.DatasetID, input.CandidateExecution)
		writeEvaluationResult(w, r, item, err, http.StatusCreated)
	}
}

func getEvaluationRun(store *evaluation.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, err := store.GetRun(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "runID"))
		writeEvaluationResult(w, r, item, err, http.StatusOK)
	}
}

func writeEvaluationResult(w http.ResponseWriter, r *http.Request, item any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, item)
	case errors.Is(err, evaluation.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, evaluation.ErrInvalidInput):
		writeError(w, r, http.StatusBadRequest, "invalid_request", "evaluation configuration is invalid")
	case errors.Is(err, evaluation.ErrInvalidTransition):
		writeError(w, r, http.StatusConflict, "invalid_transition", "evaluation run cannot transition")
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "request could not be completed")
	}
}
