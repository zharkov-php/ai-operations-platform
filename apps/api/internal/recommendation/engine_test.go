package recommendation

import (
	"testing"

	"github.com/shopspring/decimal"
)

func baseSnapshot() Snapshot {
	return Snapshot{WorkloadID: "workload", Provider: "fictional", Model: "model", TaskType: "generation", Determinism: "low", Complexity: "medium", Criticality: "medium", QualityRequirement: "standard", PrivacyClassification: "internal", FailureImpact: "medium", Currency: "USD", CallCount: 1000, HashedPromptCount: 1000, UniquePromptCount: 1000, UniqueTemplateCount: 10, AverageInputTokens: decimal.NewFromInt(1000), AverageOutputTokens: decimal.NewFromInt(200), CurrentMonthlyCost: decimal.NewFromInt(1000), PricingCompleteness: decimal.NewFromInt(1), ClassificationConfidence: decimal.RequireFromString("0.9")}
}
func containsType(drafts []Draft, kind string) bool {
	for _, draft := range drafts {
		if draft.Type == kind {
			return true
		}
	}
	return false
}
func TestEveryRecommendationRule(t *testing.T) {
	tests := []struct {
		name, kind string
		mutate     func(*Snapshot)
	}{
		{"deterministic code", "replace_with_code", func(s *Snapshot) { s.TaskType = "calculation"; s.Determinism = "high" }},
		{"exact cache", "exact_cache", func(s *Snapshot) { s.UniquePromptCount = 500 }},
		{"smaller model", "smaller_hosted_model", func(s *Snapshot) { s.TaskType = "classification"; s.Complexity = "low"; s.FailureImpact = "low" }},
		{"local model", "local_model", func(s *Snapshot) { s.TaskType = "translation"; s.PrivacyClassification = "confidential" }},
		{"context reduction", "reduce_context", func(s *Snapshot) {
			s.AverageInputTokens = decimal.NewFromInt(8000)
			s.AverageRelevantInputTokens = decimal.NewFromInt(2000)
		}},
		{"output reduction", "reduce_output", func(s *Snapshot) {
			s.AverageOutputTokens = decimal.NewFromInt(600)
			s.AverageRequiredOutputTokens = decimal.NewFromInt(200)
		}},
		{"structured output", "structured_output", func(s *Snapshot) { s.RequiresStructuredOutput = true; s.JSONParseFailures = 4 }},
		{"batching", "batch_requests", func(s *Snapshot) {
			s.MaxSimilarCallsPerMinute = 12
			s.AverageInputTokens = decimal.NewFromInt(400)
			s.AverageOutputTokens = decimal.NewFromInt(100)
		}},
		{"keep current", "keep_current_model", func(s *Snapshot) { s.Criticality = "critical" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := baseSnapshot()
			test.mutate(&snapshot)
			drafts := Analyze(snapshot)
			if !containsType(drafts, test.kind) {
				t.Fatalf("missing %s in %+v", test.kind, drafts)
			}
		})
	}
}
func TestFalsePositiveBoundaries(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.CallCount = 19
	snapshot.UniquePromptCount = 1
	if drafts := Analyze(snapshot); len(drafts) != 1 || drafts[0].Type != "keep_current_model" {
		t.Fatalf("low volume drafts=%+v", drafts)
	}
	snapshot = baseSnapshot()
	snapshot.UniquePromptCount = 801
	if containsType(Analyze(snapshot), "exact_cache") {
		t.Fatal("cache recommended below repetition threshold")
	}
	snapshot = baseSnapshot()
	snapshot.TaskType = "classification"
	snapshot.FailureImpact = "high"
	if containsType(Analyze(snapshot), "smaller_hosted_model") {
		t.Fatal("smaller model recommended for high failure impact")
	}
}
func TestConfidenceUsesMeasurableInputs(t *testing.T) {
	low := baseSnapshot()
	low.CallCount = 20
	low.UniqueTemplateCount = 20
	low.ClassificationConfidence = decimal.RequireFromString("0.2")
	low.PricingCompleteness = decimal.RequireFromString("0.5")
	high := baseSnapshot()
	lowDraft := Analyze(low)[0]
	highDraft := Analyze(high)[0]
	if !highDraft.Confidence.GreaterThan(lowDraft.Confidence) {
		t.Fatalf("low=%s high=%s", lowDraft.Confidence, highDraft.Confidence)
	}
	if highDraft.ConfidenceInputs.AnalyzedCallCount != high.CallCount {
		t.Fatal("call count missing from confidence inputs")
	}
}
func TestEconomicsNeverClaimVerifiedSavings(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.UniquePromptCount = 100
	drafts := Analyze(snapshot)
	for _, draft := range drafts {
		if _, exists := draft.Evidence["verified_savings"]; exists {
			t.Fatal("draft claimed verified savings")
		}
		if draft.EstimatedMonthlySavings.IsPositive() && draft.EstimatedBreakEvenMonths == nil {
			t.Fatal("missing break-even estimate")
		}
	}
}
