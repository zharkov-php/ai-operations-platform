package recommendation

import (
	"math"
	"sort"

	"github.com/shopspring/decimal"
)

const RuleVersion = "v1.0.0"

type Snapshot struct {
	WorkloadID, Provider, Model, TaskType, Determinism, Complexity, Criticality, QualityRequirement, PrivacyClassification, FailureImpact, Currency string
	CallCount, HashedPromptCount, UniquePromptCount, UniqueTemplateCount, JSONParseFailures, MalformedRetries                                       int64
	AverageInputTokens, AverageOutputTokens, AverageRelevantInputTokens, AverageRequiredOutputTokens                                                decimal.Decimal
	CurrentMonthlyCost                                                                                                                              decimal.Decimal
	PricingCompleteness, ClassificationConfidence                                                                                                   decimal.Decimal
	RequiresStructuredOutput, RequiresToolCalling, ExternalKnowledgeRequired, NaturalLanguageUnderstandingRequired                                  bool
	MaxSimilarCallsPerMinute                                                                                                                        int64
}
type ConfidenceInputs struct {
	AnalyzedCallCount        int64  `json:"analyzed_call_count"`
	TemplateStability        string `json:"template_stability"`
	RepeatedInputRate        string `json:"repeated_input_rate"`
	ClassificationConfidence string `json:"classification_confidence"`
	PricingCompleteness      string `json:"pricing_completeness"`
	GroundTruthAvailable     bool   `json:"ground_truth_available"`
	EvaluationAvailable      bool   `json:"evaluation_available"`
}
type Draft struct {
	Type, Priority, ConfidenceLevel                      string
	Confidence                                           decimal.Decimal
	ReasonCodes                                          []string
	Evidence                                             map[string]any
	ConfidenceInputs                                     ConfidenceInputs
	EstimatedMonthlySavings, EstimatedImplementationCost decimal.Decimal
	EstimatedBreakEvenMonths                             *decimal.Decimal
	QualityRisk, OperationalRisk, RequiredNextAction     string
	CurrentExecution, ProposedExecution                  map[string]any
}

func Analyze(s Snapshot) []Draft {
	confidence, inputs := calculateConfidence(s)
	base := func(kind, priority string) Draft {
		return Draft{Type: kind, Priority: priority, Confidence: confidence, ConfidenceLevel: level(confidence), ConfidenceInputs: inputs, CurrentExecution: map[string]any{"provider": s.Provider, "model": s.Model, "monthly_cost": s.CurrentMonthlyCost.String()}, Evidence: evidence(s), QualityRisk: "Candidate behavior may differ from the current execution.", OperationalRisk: "Implementation and rollback procedures require review.", RequiredNextAction: "Review evidence and define an evaluation before implementation."}
	}
	if s.CallCount < 20 || s.Criticality == "critical" || s.QualityRequirement == "critical" || s.FailureImpact == "critical" {
		draft := base("keep_current_model", "high")
		draft.ReasonCodes = []string{"QUALITY_OR_EVIDENCE_GUARDRAIL"}
		draft.ProposedExecution = map[string]any{"action": "keep_current_model"}
		draft.QualityRisk = "Changing execution without sufficient evidence may cause unacceptable failures."
		draft.OperationalRisk = "No change recommended."
		draft.RequiredNextAction = "Collect more representative calls and ground-truth evaluation data."
		return []Draft{withEconomics(draft, decimal.Zero, decimal.Zero)}
	}
	drafts := []Draft{}
	add := func(d Draft, savings, implementation decimal.Decimal) {
		drafts = append(drafts, withEconomics(d, savings, implementation))
	}
	if oneOf(s.TaskType, "calculation", "validation", "format_conversion") && s.Determinism == "high" && s.CallCount >= 50 {
		d := base("replace_with_code", "high")
		d.ReasonCodes = []string{"DETERMINISTIC_TASK", "MEANINGFUL_VOLUME"}
		d.ProposedExecution = map[string]any{"type": "deterministic_code", "automatic_deployment": false}
		d.QualityRisk = "Configured rules may omit edge cases present in natural-language instructions."
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.70")), decimal.NewFromInt(1200))
	}
	repeatedRate := repeatRate(s)
	if s.HashedPromptCount >= 20 && repeatedRate.GreaterThanOrEqual(decimal.RequireFromString("0.20")) {
		d := base("exact_cache", "high")
		d.ReasonCodes = []string{"REPEATED_PROMPT_HASHES", "EXACT_MATCH_ONLY"}
		d.ProposedExecution = map[string]any{"type": "exact_cache"}
		d.QualityRisk = "Stale output is possible if external state is not included in the cache key."
		add(d, s.CurrentMonthlyCost.Mul(repeatedRate).Mul(decimal.RequireFromString("0.85")), decimal.NewFromInt(400))
	}
	if oneOf(s.TaskType, "classification", "extraction") && oneOf(s.Complexity, "low", "medium") && !oneOf(s.FailureImpact, "high", "critical") && s.CallCount >= 100 && s.CurrentMonthlyCost.GreaterThanOrEqual(decimal.NewFromInt(10)) {
		d := base("smaller_hosted_model", "medium")
		d.ReasonCodes = []string{"REPETITIVE_STRUCTURED_TASK", "LOW_OR_MEDIUM_COMPLEXITY", "EVALUATION_REQUIRED"}
		d.ProposedExecution = map[string]any{"type": "smaller_hosted_model_candidate"}
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.45")), decimal.NewFromInt(600))
	}
	if oneOf(s.PrivacyClassification, "confidential", "restricted") && oneOf(s.TaskType, "translation", "summarization", "classification", "extraction") && s.QualityRequirement != "critical" && s.CallCount >= 500 && !s.RequiresToolCalling && s.AverageInputTokens.LessThanOrEqual(decimal.NewFromInt(16000)) && s.CurrentMonthlyCost.GreaterThan(decimal.NewFromInt(300)) {
		d := base("local_model", "medium")
		d.ReasonCodes = []string{"SENSITIVE_REPETITIVE_LANGUAGE_TASK", "VOLUME_SUPPORTS_ECONOMIC_REVIEW", "LOCAL_QUALITY_UNVERIFIED"}
		d.ProposedExecution = map[string]any{"type": "local_model_candidate", "illustrative_monthly_infrastructure_cost": "200", "hardware_amortization": "120", "electricity_estimate": "40", "maintenance_estimate": "40", "monthly_capacity_calls": 50000, "context_limit": 16000, "throughput": "measurement_required", "latency": "measurement_required", "model_quality": "evaluation_required"}
		d.QualityRisk = "Local model quality has not been established for this workload."
		d.OperationalRisk = "Hardware capacity, latency, maintenance, context limits, and utilization must be measured."
		savings := s.CurrentMonthlyCost.Sub(decimal.NewFromInt(200))
		add(d, savings, decimal.NewFromInt(2500))
	}
	if s.CallCount >= 50 && s.AverageInputTokens.GreaterThanOrEqual(decimal.NewFromInt(4000)) && s.AverageRelevantInputTokens.IsPositive() && s.AverageRelevantInputTokens.Div(s.AverageInputTokens).LessThan(decimal.RequireFromString("0.70")) {
		d := base("reduce_context", "medium")
		d.ReasonCodes = []string{"LARGE_INPUT", "MEASURED_LOW_RELEVANT_TOKEN_RATIO"}
		d.ProposedExecution = map[string]any{"type": "reduced_context", "semantic_relevance_claimed": false}
		d.QualityRisk = "Removing context may omit information required for uncommon cases."
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.20")), decimal.NewFromInt(500))
	}
	if s.CallCount >= 50 && s.AverageRequiredOutputTokens.IsPositive() && s.AverageOutputTokens.GreaterThan(s.AverageRequiredOutputTokens.Mul(decimal.NewFromInt(2))) {
		d := base("reduce_output", "medium")
		d.ReasonCodes = []string{"OUTPUT_EXCEEDS_CONFIGURED_REQUIREMENT"}
		d.ProposedExecution = map[string]any{"type": "reduced_output"}
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.15")), decimal.NewFromInt(250))
	}
	if s.RequiresStructuredOutput && (s.JSONParseFailures >= 3 || s.MalformedRetries >= 3) {
		d := base("structured_output", "high")
		d.ReasonCodes = []string{"KNOWN_SCHEMA", "MALFORMED_OUTPUT_FAILURES"}
		d.ProposedExecution = map[string]any{"type": "provider_structured_output"}
		d.QualityRisk = "Schema-valid output can still be semantically incorrect."
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.10")), decimal.NewFromInt(350))
	}
	if s.CallCount >= 100 && s.MaxSimilarCallsPerMinute >= 10 && s.AverageInputTokens.Add(s.AverageOutputTokens).LessThan(decimal.NewFromInt(1000)) {
		d := base("batch_requests", "low")
		d.ReasonCodes = []string{"SMALL_SIMILAR_CALL_BURSTS"}
		d.ProposedExecution = map[string]any{"type": "batch_requests"}
		d.OperationalRisk = "Batching can increase queueing latency and failure blast radius."
		add(d, s.CurrentMonthlyCost.Mul(decimal.RequireFromString("0.10")), decimal.NewFromInt(500))
	}
	if len(drafts) == 0 {
		d := base("keep_current_model", "medium")
		d.ReasonCodes = []string{"NO_RULE_THRESHOLD_MET"}
		d.ProposedExecution = map[string]any{"action": "keep_current_model"}
		d.OperationalRisk = "No change recommended."
		d.RequiredNextAction = "Continue observation and collect evaluation ground truth."
		add(d, decimal.Zero, decimal.Zero)
	}
	sort.SliceStable(drafts, func(i, j int) bool { return priorityRank(drafts[i].Priority) > priorityRank(drafts[j].Priority) })
	return drafts
}
func withEconomics(d Draft, savings, implementation decimal.Decimal) Draft {
	if savings.IsNegative() {
		savings = decimal.Zero
	}
	d.EstimatedMonthlySavings = savings
	d.EstimatedImplementationCost = implementation
	if savings.IsPositive() {
		value := implementation.Div(savings)
		d.EstimatedBreakEvenMonths = &value
	}
	return d
}
func repeatRate(s Snapshot) decimal.Decimal {
	if s.CallCount == 0 {
		return decimal.Zero
	}
	repeated := s.HashedPromptCount - s.UniquePromptCount
	if repeated < 0 {
		repeated = 0
	}
	return decimal.NewFromInt(repeated).Div(decimal.NewFromInt(s.CallCount))
}
func calculateConfidence(s Snapshot) (decimal.Decimal, ConfidenceInputs) {
	calls := math.Min(float64(s.CallCount)/500, 1)
	template := 0.0
	if s.CallCount > 0 {
		template = 1 - math.Min(float64(s.UniqueTemplateCount)/float64(s.CallCount), 1)
	}
	repeated, _ := repeatRate(s).Float64()
	classification, _ := s.ClassificationConfidence.Float64()
	pricing, _ := s.PricingCompleteness.Float64()
	numeric := decimal.NewFromFloat(calls*.25 + template*.20 + repeated*.15 + classification*.25 + pricing*.15)
	return numeric, ConfidenceInputs{AnalyzedCallCount: s.CallCount, TemplateStability: decimal.NewFromFloat(template).StringFixed(4), RepeatedInputRate: decimal.NewFromFloat(repeated).StringFixed(4), ClassificationConfidence: s.ClassificationConfidence.StringFixed(4), PricingCompleteness: s.PricingCompleteness.StringFixed(4)}
}
func level(value decimal.Decimal) string {
	if value.GreaterThanOrEqual(decimal.RequireFromString("0.75")) {
		return "high"
	}
	if value.GreaterThanOrEqual(decimal.RequireFromString("0.45")) {
		return "medium"
	}
	return "low"
}
func evidence(s Snapshot) map[string]any {
	return map[string]any{"call_count": s.CallCount, "hashed_prompt_count": s.HashedPromptCount, "unique_prompt_count": s.UniquePromptCount, "task_type": s.TaskType, "determinism": s.Determinism, "complexity": s.Complexity, "criticality": s.Criticality, "quality_requirement": s.QualityRequirement, "privacy_classification": s.PrivacyClassification, "failure_impact": s.FailureImpact, "average_input_tokens": s.AverageInputTokens.String(), "average_output_tokens": s.AverageOutputTokens.String(), "current_monthly_cost": s.CurrentMonthlyCost.String(), "pricing_completeness": s.PricingCompleteness.String(), "rule_version": RuleVersion}
}
func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func priorityRank(value string) int {
	if value == "high" {
		return 3
	}
	if value == "medium" {
		return 2
	}
	return 1
}
