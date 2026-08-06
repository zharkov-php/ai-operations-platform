package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/apikey"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
)

type createKeyRequest struct {
	Name      string   `json:"name"`
	ProjectID *string  `json:"project_id"`
	Scopes    []string `json:"scopes"`
}

func registerAPIKeyRoutes(router chi.Router, authService *auth.Service, store *apikey.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/api-keys", listAPIKeys(store))
		private.With(requireRoles("owner", "admin")).Post("/api/v1/api-keys", createAPIKey(store))
		private.With(requireRoles("owner", "admin")).Delete("/api/v1/api-keys/{keyID}", revokeAPIKey(store))
	})
}
func createAPIKey(store *apikey.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createKeyRequest
		if json.NewDecoder(r.Body).Decode(&input) != nil || apikey.Validate(input.Name, input.Scopes) != nil {
			writeError(w, r, 400, "invalid_api_key", "API key payload is invalid")
			return
		}
		claims := requestClaims(r)
		created, err := store.Create(r.Context(), claims.OrganizationID, claims.UserID, input.Name, input.ProjectID, input.Scopes)
		if err != nil {
			writeError(w, r, 400, "invalid_api_key", "API key could not be created")
			return
		}
		writeJSON(w, 201, created)
	}
}
func listAPIKeys(store *apikey.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.List(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			writeError(w, r, 500, "internal_error", "request could not be completed")
			return
		}
		writeJSON(w, 200, map[string]any{"items": items})
	}
}
func revokeAPIKey(store *apikey.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := store.Revoke(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "keyID"))
		if errors.Is(err, apikey.ErrInvalid) {
			writeError(w, r, 404, "not_found", "resource not found")
			return
		}
		if err != nil {
			writeError(w, r, 500, "internal_error", "request could not be completed")
			return
		}
		w.WriteHeader(204)
	}
}
