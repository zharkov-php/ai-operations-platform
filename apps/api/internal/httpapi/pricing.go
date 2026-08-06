package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/pricing"
)

func registerPricingRoutes(router chi.Router, authService *auth.Service, store *pricing.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/provider-models", listModels(store))
		private.Get("/api/v1/model-pricing", listPrices(store))
		private.With(requireRoles("owner", "admin")).Post("/api/v1/provider-models", createModel(store))
		private.With(requireRoles("owner", "admin")).Post("/api/v1/model-pricing", createPrice(store))
		private.With(requireRoles("owner", "admin")).Patch("/api/v1/model-pricing/{pricingID}", updatePrice(store))
	})
}
func listModels(store *pricing.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.ListModels(r.Context())
		writePricingResult(w, r, map[string]any{"items": items}, err, 200)
	}
}
func createModel(store *pricing.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input pricing.Model
		if json.NewDecoder(r.Body).Decode(&input) != nil || pricing.ValidateModel(input) != nil {
			writeError(w, r, 400, "invalid_model", "model payload is invalid")
			return
		}
		result, err := store.CreateModel(r.Context(), input)
		writePricingResult(w, r, result, err, 201)
	}
}
func listPrices(store *pricing.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		items, err := store.ListPrices(r.Context(), r.URL.Query().Get("provider"), r.URL.Query().Get("model"), r.URL.Query().Get("currency"), limit, offset)
		writePricingResult(w, r, map[string]any{"items": items}, err, 200)
	}
}
func createPrice(store *pricing.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input pricing.PriceInput
		if json.NewDecoder(r.Body).Decode(&input) != nil || pricing.Validate(input) != nil {
			writeError(w, r, 400, "invalid_pricing", "pricing payload is invalid")
			return
		}
		result, err := store.CreatePrice(r.Context(), input)
		writePricingResult(w, r, result, err, 201)
	}
}

type pricePatch struct {
	InputPrice       *string `json:"input_price"`
	OutputPrice      *string `json:"output_price"`
	CachedInputPrice *string `json:"cached_input_price"`
	EffectiveTo      *string `json:"effective_to"`
	SourceNote       *string `json:"source_note"`
	Illustrative     *bool   `json:"illustrative"`
}

func updatePrice(store *pricing.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, err := store.GetPrice(r.Context(), chi.URLParam(r, "pricingID"))
		if err != nil {
			writePricingResult(w, r, current, err, 200)
			return
		}
		var patch pricePatch
		if json.NewDecoder(r.Body).Decode(&patch) != nil {
			writeError(w, r, 400, "invalid_pricing", "pricing payload is invalid")
			return
		}
		input := pricing.PriceInput{Provider: current.Provider, Model: current.Model, InputPrice: current.InputPrice, OutputPrice: current.OutputPrice, CachedInputPrice: current.CachedInputPrice, Currency: current.Currency, UnitSize: current.UnitSize, EffectiveFrom: current.EffectiveFrom, EffectiveTo: current.EffectiveTo, SourceNote: current.SourceNote, Illustrative: current.Illustrative}
		if patch.InputPrice != nil {
			input.InputPrice = *patch.InputPrice
		}
		if patch.OutputPrice != nil {
			input.OutputPrice = *patch.OutputPrice
		}
		if patch.CachedInputPrice != nil {
			input.CachedInputPrice = patch.CachedInputPrice
		}
		if patch.SourceNote != nil {
			input.SourceNote = *patch.SourceNote
		}
		if patch.Illustrative != nil {
			input.Illustrative = *patch.Illustrative
		}
		if patch.EffectiveTo != nil {
			parsed, parseErr := time.Parse(time.RFC3339, *patch.EffectiveTo)
			if parseErr != nil {
				writeError(w, r, 400, "invalid_pricing", "effective_to is invalid")
				return
			}
			input.EffectiveTo = &parsed
		}
		if pricing.Validate(input) != nil {
			writeError(w, r, 400, "invalid_pricing", "pricing payload is invalid")
			return
		}
		result, err := store.UpdatePrice(r.Context(), current.ID, input)
		writePricingResult(w, r, result, err, 200)
	}
}
func writePricingResult(w http.ResponseWriter, r *http.Request, result any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, result)
	case errors.Is(err, pricing.ErrNotFound):
		writeError(w, r, 404, "not_found", "resource not found")
	case errors.Is(err, pricing.ErrConflict):
		writeError(w, r, 409, "pricing_period_overlap", "pricing period overlaps an existing period")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
