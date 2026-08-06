package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidInput      = errors.New("invalid evaluation input")
	ErrInvalidTransition = errors.New("invalid evaluation run transition")
	ErrNotFound          = errors.New("evaluation resource not found")
)

type Rule struct {
	Type     string            `json:"type"`
	Field    string            `json:"field,omitempty"`
	Fields   []string          `json:"fields,omitempty"`
	Required []string          `json:"required,omitempty"`
	Types    map[string]string `json:"types,omitempty"`
}

type Case struct {
	ID              string         `json:"id"`
	SanitizedInput  any            `json:"sanitized_input"`
	ExpectedOutput  any            `json:"expected_output"`
	ValidationRules []Rule         `json:"validation_rules"`
	Metadata        map[string]any `json:"metadata"`
}

type CaseResult struct {
	CaseID   string   `json:"case_id"`
	Passed   bool     `json:"passed"`
	Output   any      `json:"output"`
	Failures []string `json:"failures"`
}

type CandidateResult struct {
	Output  any
	Latency time.Duration
	Cost    decimal.Decimal
}

// CandidateAdapter is the external execution boundary. The initial adapter is
// deterministic and requires no provider credentials.
type CandidateAdapter interface {
	Execute(context.Context, Case, map[string]any) (CandidateResult, error)
}

type DeterministicAdapter struct{}

func (DeterministicAdapter) Execute(_ context.Context, item Case, config map[string]any) (CandidateResult, error) {
	mode, _ := config["mode"].(string)
	var output any
	switch mode {
	case "expected":
		output = cloneJSON(item.ExpectedOutput)
	case "echo":
		output = cloneJSON(item.SanitizedInput)
	case "fixed":
		output = cloneJSON(config["output"])
	default:
		return CandidateResult{}, fmt.Errorf("%w: unsupported deterministic mode", ErrInvalidInput)
	}
	latency, err := number(config["latency_ms"])
	if err != nil || latency.IsNegative() {
		return CandidateResult{}, fmt.Errorf("%w: latency_ms", ErrInvalidInput)
	}
	cost := decimal.Zero
	if value, ok := config["cost_per_case"]; ok {
		cost, err = number(value)
		if err != nil || cost.IsNegative() {
			return CandidateResult{}, fmt.Errorf("%w: cost_per_case", ErrInvalidInput)
		}
	}
	return CandidateResult{Output: output, Latency: time.Duration(latency.IntPart()) * time.Millisecond, Cost: cost}, nil
}

func NextStatus(current, target string) error {
	valid := (current == "pending" && target == "running") || (current == "running" && (target == "completed" || target == "failed"))
	if !valid {
		return ErrInvalidTransition
	}
	return nil
}

func ValidateRules(actual, expected any, rules []Rule) []string {
	failures := []string{}
	for _, rule := range rules {
		switch rule.Type {
		case "exact_match":
			if !reflect.DeepEqual(normalize(actual), normalize(expected)) {
				failures = append(failures, "exact_match")
			}
		case "classification_match":
			actualValue, expectedValue := valueAt(actual, rule.Field), valueAt(expected, rule.Field)
			if actualValue == nil || expectedValue == nil || !reflect.DeepEqual(actualValue, expectedValue) {
				failures = append(failures, "classification_match:"+rule.Field)
			}
		case "required_fields":
			for _, field := range rule.Fields {
				if valueAt(actual, field) == nil {
					failures = append(failures, "required_field:"+field)
				}
			}
		case "json_schema":
			for _, field := range rule.Required {
				if valueAt(actual, field) == nil {
					failures = append(failures, "schema_required:"+field)
				}
			}
			for field, expectedType := range rule.Types {
				if value := valueAt(actual, field); value != nil && jsonType(value) != expectedType {
					failures = append(failures, "schema_type:"+field)
				}
			}
		default:
			failures = append(failures, "unsupported_rule:"+rule.Type)
		}
	}
	return failures
}

func ValidateCase(item Case) error {
	if item.SanitizedInput == nil || item.ExpectedOutput == nil || len(item.ValidationRules) == 0 {
		return ErrInvalidInput
	}
	for _, rule := range item.ValidationRules {
		switch rule.Type {
		case "exact_match":
		case "classification_match":
			if strings.TrimSpace(rule.Field) == "" {
				return ErrInvalidInput
			}
		case "required_fields":
			if len(rule.Fields) == 0 {
				return ErrInvalidInput
			}
		case "json_schema":
			if len(rule.Required) == 0 && len(rule.Types) == 0 {
				return ErrInvalidInput
			}
		default:
			return ErrInvalidInput
		}
	}
	return nil
}

func valueAt(value any, path string) any {
	if path == "" {
		return normalize(value)
	}
	current := normalize(value)
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[part]
	}
	return current
}

func jsonType(value any) string {
	switch normalize(value).(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func normalize(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var normalized any
	if json.Unmarshal(data, &normalized) != nil {
		return value
	}
	return normalized
}

func cloneJSON(value any) any { return normalize(value) }

func number(value any) (decimal.Decimal, error) {
	if value == nil {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(fmt.Sprint(value))
}
