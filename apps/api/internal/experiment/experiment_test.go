package experiment

import (
	"errors"
	"reflect"
	"testing"
)

func TestExperimentTransitions(t *testing.T) {
	tests := []struct{ from, action, to string }{{"draft", "approve", "approved"}, {"approved", "start", "running"}, {"running", "pause", "paused"}, {"paused", "start", "running"}, {"running", "rollback", "rolled_back"}, {"paused", "rollback", "rolled_back"}, {"running", "complete", "completed"}, {"completed", "verify", "verified"}}
	for _, test := range tests {
		target, err := NextStatus(test.from, test.action)
		if err != nil || target != test.to {
			t.Fatalf("%s %s target=%s err=%v", test.from, test.action, target, err)
		}
	}
	if _, err := NextStatus("draft", "start"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("err=%v", err)
	}
}
func TestGuardrailViolationsAreDeterministic(t *testing.T) {
	guardrails := Guardrails{"0.90", "500", "0.05", "0.02"}
	metrics := Metrics{QualityScore: "0.80", AverageLatencyMS: "600", ErrorRate: "0.10", CostPerCall: "0.03"}
	got, err := Violations(guardrails, metrics)
	want := []string{"cost_above_maximum", "error_rate_above_maximum", "latency_above_maximum", "quality_below_minimum"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
func TestVerifiedSavingsUsesDecimalAndCanBeNegative(t *testing.T) {
	savings, err := VerifiedSavings(Metrics{ControlTotalCost: "12.345678901234", CandidateTotalCost: "2.000000000001"})
	if err != nil || savings.StringFixed(12) != "10.345678901233" {
		t.Fatalf("savings=%s err=%v", savings, err)
	}
}
