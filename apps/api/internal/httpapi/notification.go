package httpapi

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/notification"
	"net/http"
)

func registerNotificationRoutes(router chi.Router, a *auth.Service, s *notification.Store) {
	if a == nil || s == nil {
		return
	}
	router.Group(func(r chi.Router) {
		r.Use(authenticate(a))
		r.Get("/api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
			c := requestClaims(r)
			items, err := s.List(r.Context(), c.OrganizationID, c.UserID)
			if err != nil {
				writeError(w, r, 500, "internal_error", "notifications could not be loaded")
				return
			}
			writeJSON(w, 200, map[string]any{"items": items})
		})
		r.Get("/api/v1/notification-preferences", func(w http.ResponseWriter, r *http.Request) {
			p, err := s.GetPreferences(r.Context(), requestClaims(r).UserID)
			if err != nil {
				writeError(w, r, 500, "internal_error", "preferences could not be loaded")
				return
			}
			writeJSON(w, 200, p)
		})
		r.Put("/api/v1/notification-preferences", func(w http.ResponseWriter, r *http.Request) {
			var p notification.Preferences
			if json.NewDecoder(r.Body).Decode(&p) != nil {
				writeError(w, r, 400, "invalid_request", "preferences are invalid")
				return
			}
			p, err := s.SetPreferences(r.Context(), requestClaims(r).UserID, p)
			if err != nil {
				writeError(w, r, 500, "internal_error", "preferences could not be saved")
				return
			}
			writeJSON(w, 200, p)
		})
	})
}
