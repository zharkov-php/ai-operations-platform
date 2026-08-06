package evaluation

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRunStateTransitions(t *testing.T) {
	tests := []struct {
		from, to string
		valid    bool
	}{
		{"pending", "running", true}, {"running", "completed", true}, {"running", "failed", true},
		{"pending", "completed", false}, {"completed", "running", false}, {"failed", "running", false},
	}
	for _, test := range tests {
		err := NextStatus(test.from, test.to)
		if test.valid && err != nil || !test.valid && !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("%s -> %s: %v", test.from, test.to, err)
		}
	}
}

func TestValidationRulesAndPartialFailure(t *testing.T) {
	expected := map[string]any{"label": "billing", "confidence": 0.9, "details": map[string]any{"owner": "support"}}
	tests := []struct {
		name   string
		actual any
		rule   Rule
		pass   bool
	}{
		{"exact", expected, Rule{Type: "exact_match"}, true},
		{"classification", map[string]any{"label": "billing"}, Rule{Type: "classification_match", Field: "label"}, true},
		{"classification mismatch", map[string]any{"label": "technical"}, Rule{Type: "classification_match", Field: "label"}, false},
		{"required fields", expected, Rule{Type: "required_fields", Fields: []string{"label", "details.owner"}}, true},
		{"missing field", expected, Rule{Type: "required_fields", Fields: []string{"missing"}}, false},
		{"schema", expected, Rule{Type: "json_schema", Required: []string{"label"}, Types: map[string]string{"label": "string", "confidence": "number"}}, true},
		{"schema type", expected, Rule{Type: "json_schema", Types: map[string]string{"confidence": "string"}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failures := ValidateRules(test.actual, expected, []Rule{test.rule})
			if (len(failures) == 0) != test.pass {
				t.Fatalf("failures=%v", failures)
			}
		})
	}
}

func TestDeterministicAdapterRepeatabilityAndIsolation(t *testing.T) {
	item := Case{SanitizedInput: map[string]any{"ticket": "sanitized"}, ExpectedOutput: map[string]any{"label": "billing"}}
	config := map[string]any{"mode": "expected", "latency_ms": float64(12), "cost_per_case": "0.0025"}
	adapter := DeterministicAdapter{}
	first, err := adapter.Execute(context.Background(), item, config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Execute(context.Background(), item, config)
	if err != nil || !reflect.DeepEqual(first.Output, second.Output) || first.Cost.String() != "0.0025" || first.Latency.Milliseconds() != 12 {
		t.Fatalf("first=%+v second=%+v err=%v", first, second, err)
	}
	first.Output.(map[string]any)["label"] = "changed"
	if item.ExpectedOutput.(map[string]any)["label"] != "billing" {
		t.Fatal("adapter mutated the evaluation case")
	}
}

func TestCaseValidationRejectsUnsupportedOrIncompleteRules(t *testing.T) {
	base := Case{SanitizedInput: "input", ExpectedOutput: "output"}
	for _, rule := range []Rule{{Type: "unknown"}, {Type: "classification_match"}, {Type: "required_fields"}, {Type: "json_schema"}} {
		base.ValidationRules = []Rule{rule}
		if !errors.Is(ValidateCase(base), ErrInvalidInput) {
			t.Fatalf("rule should be invalid: %+v", rule)
		}
	}
}

func TestSanitizeRemovesSensitiveEvaluationContent(t *testing.T) {
	clean := sanitize(map[string]any{"password": "secret", "contact": "owner@example.test", "nested": []any{"Bearer token-value"}}).(map[string]any)
	if clean["password"] != "[REDACTED]" || clean["contact"] != "[REDACTED]" || clean["nested"].([]any)[0] != "[REDACTED]" {
		t.Fatalf("clean=%+v", clean)
	}
}
