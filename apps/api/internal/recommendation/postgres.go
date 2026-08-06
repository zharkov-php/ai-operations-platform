package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrNotFound          = errors.New("recommendation not found")
	ErrInvalidTransition = errors.New("invalid recommendation transition")
	ErrReasonRequired    = errors.New("rejection reason is required")
)

type Recommendation struct {
	ID                          string         `json:"id"`
	WorkloadID                  string         `json:"workload_id"`
	Type                        string         `json:"recommendation_type"`
	Priority                    string         `json:"priority"`
	ConfidenceLevel             string         `json:"confidence_level"`
	Status                      string         `json:"status"`
	Currency                    string         `json:"currency"`
	QualityRisk                 string         `json:"quality_risk"`
	OperationalRisk             string         `json:"operational_risk"`
	RequiredNextAction          string         `json:"required_next_action"`
	RuleVersion                 string         `json:"rule_version"`
	Confidence                  string         `json:"confidence"`
	EstimatedMonthlySavings     string         `json:"estimated_monthly_savings"`
	EstimatedImplementationCost string         `json:"estimated_implementation_cost"`
	EstimatedBreakEvenMonths    *string        `json:"estimated_break_even_months"`
	ReasonCodes                 []string       `json:"reason_codes"`
	EvidenceSummary             map[string]any `json:"evidence_summary"`
	ConfidenceInputs            map[string]any `json:"confidence_inputs"`
	CurrentExecution            map[string]any `json:"current_execution"`
	ProposedExecution           map[string]any `json:"proposed_execution"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
}
type Store struct{ pool *pgxpool.Pool }

type AuditEntry struct {
	ID          string         `json:"id"`
	ActorUserID string         `json:"actor_user_id"`
	Action      string         `json:"action"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) AnalyzeOrganization(ctx context.Context, orgID string) ([]Recommendation, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT w.id FROM workloads w JOIN projects p ON p.id=w.project_id JOIN llm_calls c ON c.workload_id=w.id WHERE p.organization_id=$1 AND w.status='active' ORDER BY w.id LIMIT 500`, orgID)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	results := []Recommendation{}
	for _, id := range ids {
		snapshot, buildErr := s.snapshot(ctx, orgID, id)
		if buildErr != nil {
			return nil, buildErr
		}
		drafts := Analyze(snapshot)
		if err = s.saveClassification(ctx, snapshot); err != nil {
			return nil, err
		}
		for _, draft := range drafts {
			saved, saveErr := s.saveRecommendation(ctx, id, snapshot.Currency, draft)
			if saveErr != nil {
				return nil, saveErr
			}
			results = append(results, saved)
		}
	}
	return results, nil
}
func (s *Store) snapshot(ctx context.Context, orgID, workloadID string) (Snapshot, error) {
	var out Snapshot
	out.WorkloadID = workloadID
	err := s.pool.QueryRow(ctx, `SELECT w.criticality,w.quality_requirement,w.privacy_classification,p.currency FROM workloads w JOIN projects p ON p.id=w.project_id WHERE p.organization_id=$1 AND w.id=$2`, orgID, workloadID).Scan(&out.Criticality, &out.QualityRequirement, &out.PrivacyClassification, &out.Currency)
	if err != nil {
		return Snapshot{}, err
	}
	var avgInput, avgOutput, avgRelevant, avgRequired, cost, classificationConfidence, pricingCompleteness string
	err = s.pool.QueryRow(ctx, `SELECT count(*),count(prompt_hash),count(DISTINCT prompt_hash),count(DISTINCT NULLIF(prompt_template_id,'')),count(*) FILTER(WHERE error_code IN('json_parse_error','malformed_output')),COALESCE(sum(retry_count) FILTER(WHERE error_code IN('json_parse_error','malformed_output')),0),COALESCE(avg(input_tokens),0)::text,COALESCE(avg(output_tokens),0)::text,COALESCE(avg(CASE WHEN (metadata->>'relevant_input_tokens') ~ '^[0-9]+$' THEN (metadata->>'relevant_input_tokens')::numeric END),0)::text,COALESCE(avg(CASE WHEN (metadata->>'required_output_tokens') ~ '^[0-9]+$' THEN (metadata->>'required_output_tokens')::numeric END),0)::text,COALESCE(sum(estimated_cost) FILTER(WHERE request_timestamp>=CURRENT_TIMESTAMP-interval '30 days'),0)::text,COALESCE(count(*) FILTER(WHERE metadata->>'task_type' IS NOT NULL)::numeric/NULLIF(count(*),0),0)::text,COALESCE(count(estimated_cost)::numeric/NULLIF(count(*),0),0)::text,COALESCE(bool_or(metadata->>'requires_structured_output'='true'),false),COALESCE(bool_or(metadata->>'requires_tool_calling'='true'),false),COALESCE(bool_or(metadata->>'external_knowledge_required'='true'),false),COALESCE(bool_or(metadata->>'natural_language_understanding_required'='true'),false) FROM llm_calls WHERE organization_id=$1 AND workload_id=$2`, orgID, workloadID).Scan(&out.CallCount, &out.HashedPromptCount, &out.UniquePromptCount, &out.UniqueTemplateCount, &out.JSONParseFailures, &out.MalformedRetries, &avgInput, &avgOutput, &avgRelevant, &avgRequired, &cost, &classificationConfidence, &pricingCompleteness, &out.RequiresStructuredOutput, &out.RequiresToolCalling, &out.ExternalKnowledgeRequired, &out.NaturalLanguageUnderstandingRequired)
	if err != nil {
		return Snapshot{}, err
	}
	out.AverageInputTokens = mustDecimal(avgInput)
	out.AverageOutputTokens = mustDecimal(avgOutput)
	out.AverageRelevantInputTokens = mustDecimal(avgRelevant)
	out.AverageRequiredOutputTokens = mustDecimal(avgRequired)
	out.CurrentMonthlyCost = mustDecimal(cost)
	out.ClassificationConfidence = mustDecimal(classificationConfidence)
	out.PricingCompleteness = mustDecimal(pricingCompleteness)
	if err = s.pool.QueryRow(ctx, `SELECT provider,model FROM llm_calls WHERE organization_id=$1 AND workload_id=$2 GROUP BY provider,model ORDER BY count(*) DESC,provider,model LIMIT 1`, orgID, workloadID).Scan(&out.Provider, &out.Model); err != nil {
		return Snapshot{}, err
	}
	if err = s.pool.QueryRow(ctx, `SELECT COALESCE(NULLIF(metadata->>'task_type',''),'unknown'),COALESCE(NULLIF(metadata->>'determinism',''),'unknown'),COALESCE(NULLIF(metadata->>'complexity',''),'unknown'),COALESCE(NULLIF(metadata->>'failure_impact',''),'unknown') FROM llm_calls WHERE organization_id=$1 AND workload_id=$2 GROUP BY 1,2,3,4 ORDER BY count(*) DESC LIMIT 1`, orgID, workloadID).Scan(&out.TaskType, &out.Determinism, &out.Complexity, &out.FailureImpact); err != nil {
		return Snapshot{}, err
	}
	out.TaskType = validOr(out.TaskType, "unknown", []string{"calculation", "validation", "classification", "extraction", "translation", "summarization", "generation", "reasoning", "coding", "search", "tool_selection", "format_conversion"})
	out.Determinism = validOr(out.Determinism, "unknown", []string{"low", "medium", "high"})
	out.Complexity = validOr(out.Complexity, "unknown", []string{"low", "medium", "high"})
	out.FailureImpact = validOr(out.FailureImpact, out.Criticality, []string{"low", "medium", "high", "critical"})
	if err = s.pool.QueryRow(ctx, `SELECT COALESCE(max(bucket_count),0) FROM (SELECT count(*) bucket_count FROM llm_calls WHERE organization_id=$1 AND workload_id=$2 GROUP BY date_trunc('minute',request_timestamp),provider,model) b`, orgID, workloadID).Scan(&out.MaxSimilarCallsPerMinute); err != nil {
		return Snapshot{}, err
	}
	return out, nil
}
func validOr(value, fallback string, allowed []string) string {
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}
func mustDecimal(value string) decimal.Decimal {
	result, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero
	}
	return result
}
func (s *Store) saveClassification(ctx context.Context, snapshot Snapshot) error {
	evidence, _ := json.Marshal(evidence(snapshot))
	_, err := s.pool.Exec(ctx, `INSERT INTO task_classifications(workload_id,task_type,determinism,complexity,privacy_classification,quality_requirement,failure_impact,external_knowledge_required,natural_language_understanding_required,confidence,classification_source,evidence) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'deterministic_metadata_v1',$11) ON CONFLICT(workload_id,classification_source) DO UPDATE SET task_type=EXCLUDED.task_type,determinism=EXCLUDED.determinism,complexity=EXCLUDED.complexity,failure_impact=EXCLUDED.failure_impact,confidence=EXCLUDED.confidence,evidence=EXCLUDED.evidence`, snapshot.WorkloadID, snapshot.TaskType, snapshot.Determinism, snapshot.Complexity, snapshot.PrivacyClassification, snapshot.QualityRequirement, snapshot.FailureImpact, snapshot.ExternalKnowledgeRequired, snapshot.NaturalLanguageUnderstandingRequired, snapshot.ClassificationConfidence, evidence)
	return err
}
func (s *Store) saveRecommendation(ctx context.Context, workloadID, currency string, draft Draft) (Recommendation, error) {
	current, _ := json.Marshal(draft.CurrentExecution)
	proposed, _ := json.Marshal(draft.ProposedExecution)
	summary, _ := json.Marshal(draft.Evidence)
	inputs, _ := json.Marshal(draft.ConfidenceInputs)
	var breakEven any
	if draft.EstimatedBreakEvenMonths != nil {
		breakEven = draft.EstimatedBreakEvenMonths.StringFixed(4)
	}
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO recommendations(workload_id,recommendation_type,priority,confidence_level,confidence,current_execution,proposed_execution,reason_codes,evidence_summary,confidence_inputs,estimated_monthly_savings,currency,estimated_implementation_cost,estimated_break_even_months,quality_risk,operational_risk,required_next_action,rule_version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) ON CONFLICT(workload_id,recommendation_type,rule_version) DO UPDATE SET priority=EXCLUDED.priority,confidence_level=EXCLUDED.confidence_level,confidence=EXCLUDED.confidence,current_execution=EXCLUDED.current_execution,proposed_execution=EXCLUDED.proposed_execution,reason_codes=EXCLUDED.reason_codes,evidence_summary=EXCLUDED.evidence_summary,confidence_inputs=EXCLUDED.confidence_inputs,estimated_monthly_savings=EXCLUDED.estimated_monthly_savings,currency=EXCLUDED.currency,estimated_implementation_cost=EXCLUDED.estimated_implementation_cost,estimated_break_even_months=EXCLUDED.estimated_break_even_months,quality_risk=EXCLUDED.quality_risk,operational_risk=EXCLUDED.operational_risk,required_next_action=EXCLUDED.required_next_action,updated_at=CURRENT_TIMESTAMP RETURNING id`, workloadID, draft.Type, draft.Priority, draft.ConfidenceLevel, draft.Confidence.StringFixed(4), current, proposed, draft.ReasonCodes, summary, inputs, draft.EstimatedMonthlySavings.StringFixed(12), currency, draft.EstimatedImplementationCost.StringFixed(12), breakEven, draft.QualityRisk, draft.OperationalRisk, draft.RequiredNextAction, RuleVersion).Scan(&id)
	if err != nil {
		return Recommendation{}, err
	}
	return s.Get(ctx, "", id)
}
func (s *Store) List(ctx context.Context, orgID, status, priority string) ([]Recommendation, error) {
	rows, err := s.pool.Query(ctx, baseSelect+` WHERE p.organization_id=$1 AND ($2='' OR r.status=$2) AND ($3='' OR r.priority=$3) ORDER BY CASE r.priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END DESC,r.created_at DESC LIMIT 100`, orgID, status, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Recommendation{}
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Get(ctx context.Context, orgID, id string) (Recommendation, error) {
	query := baseSelect + ` WHERE r.id=$1 AND (NULLIF($2,'')::uuid IS NULL OR p.organization_id=NULLIF($2,'')::uuid)`
	item, err := scan(s.pool.QueryRow(ctx, query, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Recommendation{}, ErrNotFound
	}
	return item, err
}

func NextReviewStatus(current, action string) (string, error) {
	if current != "new" && current != "under_review" {
		return "", ErrInvalidTransition
	}
	switch action {
	case "accept":
		return "accepted", nil
	case "reject":
		return "rejected", nil
	default:
		return "", ErrInvalidTransition
	}
}

func (s *Store) Review(ctx context.Context, orgID, actorID, id, action, reason string) (Recommendation, error) {
	reason = strings.TrimSpace(reason)
	if action == "reject" && reason == "" {
		return Recommendation{}, ErrReasonRequired
	}
	if utf8.RuneCountInString(reason) > 1000 {
		return Recommendation{}, ErrReasonRequired
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Recommendation{}, err
	}
	defer tx.Rollback(ctx)
	var current string
	err = tx.QueryRow(ctx, `SELECT r.status FROM recommendations r JOIN workloads w ON w.id=r.workload_id JOIN projects p ON p.id=w.project_id WHERE r.id=$1 AND p.organization_id=$2 FOR UPDATE OF r`, id, orgID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return Recommendation{}, ErrNotFound
	}
	if err != nil {
		return Recommendation{}, err
	}
	target, err := NextReviewStatus(current, action)
	if err != nil {
		return Recommendation{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE recommendations SET status=$2,updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, target); err != nil {
		return Recommendation{}, err
	}
	metadata, _ := json.Marshal(map[string]any{"from_status": current, "to_status": target, "reason": reason, "effect": "authorizes_evaluation_only"})
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id,metadata) VALUES($1,$2,$3,'recommendation',$4,$5)`, orgID, actorID, "recommendation."+target, id, metadata); err != nil {
		return Recommendation{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Recommendation{}, err
	}
	return s.Get(ctx, orgID, id)
}

func (s *Store) AuditHistory(ctx context.Context, orgID, id string) ([]AuditEntry, error) {
	if _, err := s.Get(ctx, orgID, id); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id,actor_user_id,action,metadata,created_at FROM audit_entries WHERE organization_id=$1 AND subject_type='recommendation' AND subject_id=$2 ORDER BY created_at,id`, orgID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AuditEntry{}
	for rows.Next() {
		var item AuditEntry
		if err = rows.Scan(&item.ID, &item.ActorUserID, &item.Action, &item.Metadata, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

const baseSelect = `SELECT r.id,r.workload_id,r.recommendation_type,r.priority,r.confidence_level,r.confidence::text,r.status,r.current_execution,r.proposed_execution,r.reason_codes,r.evidence_summary,r.confidence_inputs,r.estimated_monthly_savings::text,r.currency,r.estimated_implementation_cost::text,r.estimated_break_even_months::text,r.quality_risk,r.operational_risk,r.required_next_action,r.rule_version,r.created_at,r.updated_at FROM recommendations r JOIN workloads w ON w.id=r.workload_id JOIN projects p ON p.id=w.project_id`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Recommendation, error) {
	var r Recommendation
	err := row.Scan(&r.ID, &r.WorkloadID, &r.Type, &r.Priority, &r.ConfidenceLevel, &r.Confidence, &r.Status, &r.CurrentExecution, &r.ProposedExecution, &r.ReasonCodes, &r.EvidenceSummary, &r.ConfidenceInputs, &r.EstimatedMonthlySavings, &r.Currency, &r.EstimatedImplementationCost, &r.EstimatedBreakEvenMonths, &r.QualityRisk, &r.OperationalRisk, &r.RequiredNextAction, &r.RuleVersion, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}
