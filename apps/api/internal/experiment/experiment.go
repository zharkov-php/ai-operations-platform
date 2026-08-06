package experiment

import (
	"errors"
	"sort"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidInput      = errors.New("invalid experiment input")
	ErrInvalidTransition = errors.New("invalid experiment transition")
	ErrNotFound          = errors.New("experiment not found")
)

type Guardrails struct {
	MinimumQualityScore string `json:"minimum_quality_score"`
	MaximumLatencyMS    string `json:"maximum_latency_ms"`
	MaximumErrorRate    string `json:"maximum_error_rate"`
	MaximumCostPerCall  string `json:"maximum_cost_per_call"`
}

type Metrics struct {
	QualityScore       string `json:"quality_score"`
	AverageLatencyMS   string `json:"average_latency_ms"`
	ErrorRate          string `json:"error_rate"`
	CostPerCall        string `json:"cost_per_call"`
	ControlTotalCost   string `json:"control_total_cost"`
	CandidateTotalCost string `json:"candidate_total_cost"`
}

func NextStatus(current, action string) (string, error) {
	transitions := map[string]map[string]string{
		"draft": {"approve": "approved"}, "approved": {"start": "running"},
		"running": {"pause": "paused", "rollback": "rolled_back", "complete": "completed"},
		"paused":  {"start": "running", "rollback": "rolled_back"}, "completed": {"verify": "verified"},
	}
	if target := transitions[current][action]; target != "" {
		return target, nil
	}
	return "", ErrInvalidTransition
}

func Violations(guardrails Guardrails, metrics Metrics) ([]string, error) {
	minimumQuality, err := ratio(guardrails.MinimumQualityScore)
	if err != nil {
		return nil, err
	}
	maximumLatency, err := nonNegative(guardrails.MaximumLatencyMS)
	if err != nil {
		return nil, err
	}
	maximumErrors, err := ratio(guardrails.MaximumErrorRate)
	if err != nil {
		return nil, err
	}
	maximumCost, err := nonNegative(guardrails.MaximumCostPerCall)
	if err != nil {
		return nil, err
	}
	quality, err := ratio(metrics.QualityScore)
	if err != nil {
		return nil, err
	}
	latency, err := nonNegative(metrics.AverageLatencyMS)
	if err != nil {
		return nil, err
	}
	errors, err := ratio(metrics.ErrorRate)
	if err != nil {
		return nil, err
	}
	cost, err := nonNegative(metrics.CostPerCall)
	if err != nil {
		return nil, err
	}
	violations := []string{}
	if quality.LessThan(minimumQuality) {
		violations = append(violations, "quality_below_minimum")
	}
	if latency.GreaterThan(maximumLatency) {
		violations = append(violations, "latency_above_maximum")
	}
	if errors.GreaterThan(maximumErrors) {
		violations = append(violations, "error_rate_above_maximum")
	}
	if cost.GreaterThan(maximumCost) {
		violations = append(violations, "cost_above_maximum")
	}
	sort.Strings(violations)
	return violations, nil
}

func VerifiedSavings(metrics Metrics) (decimal.Decimal, error) {
	control, err := nonNegative(metrics.ControlTotalCost)
	if err != nil {
		return decimal.Zero, err
	}
	candidate, err := nonNegative(metrics.CandidateTotalCost)
	if err != nil {
		return decimal.Zero, err
	}
	return control.Sub(candidate), nil
}

func nonNegative(value string) (decimal.Decimal, error) {
	result, err := decimal.NewFromString(value)
	if err != nil || result.IsNegative() {
		return decimal.Zero, ErrInvalidInput
	}
	return result, nil
}
func ratio(value string) (decimal.Decimal, error) {
	result, err := nonNegative(value)
	if err != nil || result.GreaterThan(decimal.NewFromInt(1)) {
		return decimal.Zero, ErrInvalidInput
	}
	return result, nil
}
