package localmodel

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidInput = errors.New("invalid local model input")
	ErrNotFound     = errors.New("local model configuration not found")
)

var secondsPerMonth = decimal.NewFromInt(30 * 24 * 60 * 60)

type Configuration struct {
	ID                         string `json:"id"`
	HardwareName               string `json:"hardware_name"`
	PurchaseCost               string `json:"purchase_cost"`
	UsefulLifetimeMonths       int    `json:"useful_lifetime_months"`
	MonthlyElectricity         string `json:"monthly_electricity"`
	MonthlyMaintenance         string `json:"monthly_maintenance"`
	AvailableMemoryGB          string `json:"available_memory_gb"`
	EstimatedRequestsPerSecond string `json:"estimated_requests_per_second"`
	Utilization                string `json:"utilization"`
	SupportedModel             string `json:"supported_model"`
	ContextLimit               int    `json:"context_limit"`
	Currency                   string `json:"currency"`
	BenchmarkSource            string `json:"benchmark_source"`
}

type ComparisonInput struct {
	MonthlyRequests       int64  `json:"monthly_requests"`
	HostedCostPerRequest  string `json:"hosted_cost_per_request"`
	RequiredMemoryGB      string `json:"required_memory_gb"`
	RequiredContextTokens int    `json:"required_context_tokens"`
}

type Comparison struct {
	Supported                bool     `json:"supported"`
	Constraints              []string `json:"constraints"`
	MonthlyCapacity          string   `json:"monthly_capacity"`
	HardwareAmortization     string   `json:"hardware_amortization"`
	LocalMonthlyCost         string   `json:"local_monthly_cost"`
	HostedMonthlyCost        string   `json:"hosted_monthly_cost"`
	LocalCostPerRequest      *string  `json:"local_cost_per_request"`
	EstimatedMonthlySavings  string   `json:"estimated_monthly_savings"`
	EstimatedBreakEvenMonths *string  `json:"estimated_break_even_months"`
	Currency                 string   `json:"currency"`
	Methodology              string   `json:"methodology"`
}

func Compare(config Configuration, input ComparisonInput) (Comparison, error) {
	purchase, err := positiveOrZero(config.PurchaseCost)
	if err != nil || config.UsefulLifetimeMonths <= 0 || input.MonthlyRequests < 0 || len(config.Currency) != 3 {
		return Comparison{}, ErrInvalidInput
	}
	electricity, err := positiveOrZero(config.MonthlyElectricity)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}
	maintenance, err := positiveOrZero(config.MonthlyMaintenance)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}
	memory, err := positive(config.AvailableMemoryGB)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}
	throughput, err := positive(config.EstimatedRequestsPerSecond)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}
	utilization, err := positive(config.Utilization)
	if err != nil || utilization.GreaterThan(decimal.NewFromInt(1)) {
		return Comparison{}, ErrInvalidInput
	}
	hostedRate, err := positiveOrZero(input.HostedCostPerRequest)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}
	requiredMemory, err := positiveOrZero(input.RequiredMemoryGB)
	if err != nil {
		return Comparison{}, ErrInvalidInput
	}

	amortization := purchase.Div(decimal.NewFromInt(int64(config.UsefulLifetimeMonths)))
	localMonthly := amortization.Add(electricity).Add(maintenance)
	capacity := throughput.Mul(secondsPerMonth).Mul(utilization).Floor()
	hostedMonthly := hostedRate.Mul(decimal.NewFromInt(input.MonthlyRequests))
	constraints := []string{}
	if decimal.NewFromInt(input.MonthlyRequests).GreaterThan(capacity) {
		constraints = append(constraints, "monthly_volume_exceeds_capacity")
	}
	if requiredMemory.GreaterThan(memory) {
		constraints = append(constraints, "required_memory_exceeds_available_memory")
	}
	if input.RequiredContextTokens > config.ContextLimit {
		constraints = append(constraints, "required_context_exceeds_model_limit")
	}
	if input.MonthlyRequests == 0 {
		constraints = append(constraints, "monthly_volume_is_zero")
	}
	supported := len(constraints) == 0
	savings := hostedMonthly.Sub(localMonthly)
	var perRequest, breakEven *string
	if input.MonthlyRequests > 0 {
		value := localMonthly.Div(decimal.NewFromInt(input.MonthlyRequests)).StringFixed(12)
		perRequest = &value
	}
	operatingSavings := hostedMonthly.Sub(electricity).Sub(maintenance)
	if supported && purchase.IsPositive() && operatingSavings.IsPositive() {
		value := purchase.Div(operatingSavings).StringFixed(4)
		breakEven = &value
	}
	return Comparison{Supported: supported, Constraints: constraints, MonthlyCapacity: capacity.StringFixed(0), HardwareAmortization: amortization.StringFixed(12), LocalMonthlyCost: localMonthly.StringFixed(12), HostedMonthlyCost: hostedMonthly.StringFixed(12), LocalCostPerRequest: perRequest, EstimatedMonthlySavings: savings.StringFixed(12), EstimatedBreakEvenMonths: breakEven, Currency: config.Currency, Methodology: "30-day capacity; straight-line hardware amortization; break-even uses purchase cost divided by hosted cost minus monthly electricity and maintenance"}, nil
}

func positive(value string) (decimal.Decimal, error) {
	result, err := decimal.NewFromString(value)
	if err != nil || !result.IsPositive() {
		return decimal.Zero, ErrInvalidInput
	}
	return result, nil
}

func positiveOrZero(value string) (decimal.Decimal, error) {
	result, err := decimal.NewFromString(value)
	if err != nil || result.IsNegative() {
		return decimal.Zero, ErrInvalidInput
	}
	return result, nil
}
