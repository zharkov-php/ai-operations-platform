package httpapi

import (
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/analytics"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func registerAnalyticsRoutes(router chi.Router, authService *auth.Service, store *analytics.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/analytics/overview", analyticsOverview(store))
		private.Get("/api/v1/analytics/costs", analyticsCosts(store))
		private.Get("/api/v1/analytics/tokens", analyticsTokens(store))
		private.Get("/api/v1/analytics/latency", analyticsLatency(store))
		private.Get("/api/v1/analytics/errors", analyticsErrors(store))
	})
}
func analyticsFilters(r *http.Request, store *analytics.Store) (analytics.Filters, error) {
	query := r.URL.Query()
	dateRange, err := analytics.ParseRange(query.Get("from"), query.Get("to"), query.Get("timezone"), time.Now())
	if err != nil {
		return analytics.Filters{}, err
	}
	for _, id := range []string{query.Get("project_id"), query.Get("workload_id")} {
		if id != "" && !uuidPattern.MatchString(id) {
			return analytics.Filters{}, analytics.ErrInvalidRange
		}
	}
	currency := query.Get("currency")
	if currency == "" {
		currency, err = store.DefaultCurrency(r.Context(), requestClaims(r).OrganizationID)
		if err != nil {
			return analytics.Filters{}, err
		}
	}
	if !regexp.MustCompile(`^[A-Za-z]{3}$`).MatchString(currency) {
		return analytics.Filters{}, analytics.ErrInvalidRange
	}
	return analytics.Filters{Range: dateRange, ProjectID: query.Get("project_id"), WorkloadID: query.Get("workload_id"), Provider: query.Get("provider"), Model: query.Get("model"), Currency: currency}, nil
}
func analyticsOverview(store *analytics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters, err := analyticsFilters(r, store)
		if err != nil {
			writeError(w, r, 400, "invalid_range", "analytics filters are invalid")
			return
		}
		result, err := store.Overview(r.Context(), requestClaims(r).OrganizationID, filters)
		writeAnalytics(w, r, result, err)
	}
}
func analyticsCosts(store *analytics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters, err := analyticsFilters(r, store)
		if err != nil {
			writeError(w, r, 400, "invalid_range", "analytics filters are invalid")
			return
		}
		dimension := r.URL.Query().Get("dimension")
		if dimension == "" {
			dimension = "project"
		}
		breakdown, err := store.Costs(r.Context(), requestClaims(r).OrganizationID, dimension, filters)
		if err != nil {
			writeError(w, r, 400, "invalid_dimension", "cost dimension is invalid")
			return
		}
		daily, err := store.Daily(r.Context(), requestClaims(r).OrganizationID, filters)
		writeAnalytics(w, r, map[string]any{"dimension": dimension, "breakdown": breakdown, "daily": daily, "currency": filters.Currency}, err)
	}
}
func analyticsTokens(store *analytics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters, err := analyticsFilters(r, store)
		if err != nil {
			writeError(w, r, 400, "invalid_range", "analytics filters are invalid")
			return
		}
		result, err := store.Tokens(r.Context(), requestClaims(r).OrganizationID, filters)
		writeAnalytics(w, r, result, err)
	}
}
func analyticsLatency(store *analytics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters, err := analyticsFilters(r, store)
		if err != nil {
			writeError(w, r, 400, "invalid_range", "analytics filters are invalid")
			return
		}
		result, err := store.Latency(r.Context(), requestClaims(r).OrganizationID, filters)
		writeAnalytics(w, r, result, err)
	}
}
func analyticsErrors(store *analytics.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters, err := analyticsFilters(r, store)
		if err != nil {
			writeError(w, r, 400, "invalid_range", "analytics filters are invalid")
			return
		}
		result, err := store.Errors(r.Context(), requestClaims(r).OrganizationID, filters)
		writeAnalytics(w, r, result, err)
	}
}
func writeAnalytics(w http.ResponseWriter, r *http.Request, result any, err error) {
	if err != nil {
		writeError(w, r, 500, "analytics_failed", "analytics query could not be completed")
		return
	}
	writeJSON(w, 200, result)
}
