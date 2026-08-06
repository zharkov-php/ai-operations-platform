package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/localmodel"
)

func registerLocalModelRoutes(router chi.Router, authService *auth.Service, store *localmodel.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/local-model-configurations", func(w http.ResponseWriter, r *http.Request) {
			items, err := store.List(r.Context(), requestClaims(r).OrganizationID)
			if err != nil {
				writeError(w, r, 500, "internal_error", "request could not be completed")
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/local-model-configurations", createLocalModelConfiguration(store))
		private.With(requireRoles("owner", "admin", "engineer", "finance")).Post("/api/v1/local-model-configurations/{configurationID}/compare", compareLocalModel(store))
	})
}

func createLocalModelConfiguration(store *localmodel.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input localmodel.Configuration
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "local model configuration is invalid")
			return
		}
		claims := requestClaims(r)
		item, err := store.Create(r.Context(), claims.OrganizationID, claims.UserID, input)
		writeLocalModelResult(w, r, item, err, 201)
	}
}

func compareLocalModel(store *localmodel.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input localmodel.ComparisonInput
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			writeError(w, r, 400, "invalid_request", "comparison input is invalid")
			return
		}
		config, err := store.Get(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "configurationID"))
		if err != nil {
			writeLocalModelResult(w, r, config, err, 200)
			return
		}
		result, err := localmodel.Compare(config.Configuration, input)
		writeLocalModelResult(w, r, result, err, 200)
	}
}

func writeLocalModelResult(w http.ResponseWriter, r *http.Request, result any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, result)
	case errors.Is(err, localmodel.ErrInvalidInput):
		writeError(w, r, 400, "invalid_request", "local model economics input is invalid")
	case errors.Is(err, localmodel.ErrNotFound):
		writeError(w, r, 404, "not_found", "resource not found")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
