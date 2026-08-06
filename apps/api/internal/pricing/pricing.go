package pricing

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var ErrNotFound = errors.New("pricing not found")
var ErrConflict = errors.New("pricing period overlaps")

type Model struct {
	Provider                 string `json:"provider"`
	Model                    string `json:"model"`
	ContextWindow            int64  `json:"context_window"`
	SupportsStructuredOutput bool   `json:"supports_structured_output"`
	SupportsToolCalling      bool   `json:"supports_tool_calling"`
	SupportsCaching          bool   `json:"supports_caching"`
	DeploymentType           string `json:"deployment_type"`
	Status                   string `json:"status"`
}
type Price struct {
	ID               string     `json:"id"`
	Provider         string     `json:"provider"`
	Model            string     `json:"model"`
	InputPrice       string     `json:"input_price"`
	OutputPrice      string     `json:"output_price"`
	CachedInputPrice *string    `json:"cached_input_price"`
	Currency         string     `json:"currency"`
	UnitSize         int64      `json:"unit_size"`
	EffectiveFrom    time.Time  `json:"effective_from"`
	EffectiveTo      *time.Time `json:"effective_to"`
	SourceNote       string     `json:"source_note"`
	Illustrative     bool       `json:"illustrative"`
	CreatedAt        time.Time  `json:"created_at"`
}
type PriceInput struct {
	Provider         string     `json:"provider"`
	Model            string     `json:"model"`
	InputPrice       string     `json:"input_price"`
	OutputPrice      string     `json:"output_price"`
	CachedInputPrice *string    `json:"cached_input_price"`
	Currency         string     `json:"currency"`
	UnitSize         int64      `json:"unit_size"`
	EffectiveFrom    time.Time  `json:"effective_from"`
	EffectiveTo      *time.Time `json:"effective_to"`
	SourceNote       string     `json:"source_note"`
	Illustrative     bool       `json:"illustrative"`
}
type Usage struct{ InputTokens, OutputTokens, CachedInputTokens int64 }
type Cost struct{ Input, Output, CachedInput, Total decimal.Decimal }

func Validate(input PriceInput) error {
	if strings.TrimSpace(input.Provider) == "" || strings.TrimSpace(input.Model) == "" || !regexp.MustCompile(`^[A-Za-z]{3}$`).MatchString(input.Currency) || input.UnitSize <= 0 || input.EffectiveFrom.IsZero() || strings.TrimSpace(input.SourceNote) == "" {
		return errors.New("required pricing fields are invalid")
	}
	for _, value := range []string{input.InputPrice, input.OutputPrice} {
		amount, err := decimal.NewFromString(value)
		if err != nil || amount.IsNegative() {
			return errors.New("prices must be non-negative decimals")
		}
	}
	if input.CachedInputPrice != nil {
		amount, err := decimal.NewFromString(*input.CachedInputPrice)
		if err != nil || amount.IsNegative() {
			return errors.New("cached input price must be a non-negative decimal")
		}
	}
	if input.EffectiveTo != nil && !input.EffectiveTo.After(input.EffectiveFrom) {
		return errors.New("effective_to must be after effective_from")
	}
	return nil
}
func ValidateModel(model Model) error {
	if strings.TrimSpace(model.Provider) == "" || strings.TrimSpace(model.Model) == "" || model.ContextWindow <= 0 || !oneOf(model.DeploymentType, "hosted", "local", "private_hosted") || !oneOf(model.Status, "active", "deprecated") {
		return errors.New("invalid model")
	}
	return nil
}
func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func Calculate(price Price, usage Usage) (Cost, error) {
	if usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.CachedInputTokens < 0 {
		return Cost{}, errors.New("token counts must be non-negative")
	}
	input, err := decimal.NewFromString(price.InputPrice)
	if err != nil {
		return Cost{}, err
	}
	output, err := decimal.NewFromString(price.OutputPrice)
	if err != nil {
		return Cost{}, err
	}
	unit := decimal.NewFromInt(price.UnitSize)
	result := Cost{Input: input.Mul(decimal.NewFromInt(usage.InputTokens)).Div(unit), Output: output.Mul(decimal.NewFromInt(usage.OutputTokens)).Div(unit)}
	if usage.CachedInputTokens > 0 {
		if price.CachedInputPrice == nil {
			return Cost{}, errors.New("cached input pricing is unavailable")
		}
		cached, parseErr := decimal.NewFromString(*price.CachedInputPrice)
		if parseErr != nil {
			return Cost{}, parseErr
		}
		result.CachedInput = cached.Mul(decimal.NewFromInt(usage.CachedInputTokens)).Div(unit)
	}
	result.Total = result.Input.Add(result.Output).Add(result.CachedInput)
	return result, nil
}
