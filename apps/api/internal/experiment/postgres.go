package experiment

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Experiment struct {
	ID, WorkloadID, RecommendationID, Status, Currency, RollbackReason string
	ControlExecution, CandidateExecution, Results                      map[string]any
	TrafficPercentage, VerifiedSavings                                 *string
	Guardrails                                                         Guardrails
	StartedAt, CompletedAt                                             *time.Time
	Version                                                            int
	CreatedAt, UpdatedAt                                               time.Time
}

type CreateInput struct {
	WorkloadID         string         `json:"workload_id"`
	RecommendationID   string         `json:"recommendation_id"`
	TrafficPercentage  string         `json:"traffic_percentage"`
	Currency           string         `json:"currency"`
	ControlExecution   map[string]any `json:"control_execution"`
	CandidateExecution map[string]any `json:"candidate_execution"`
	Guardrails         Guardrails     `json:"guardrails"`
}

func (e Experiment) MarshalJSON() ([]byte, error) {
	type output struct {
		ID                 string         `json:"id"`
		WorkloadID         string         `json:"workload_id"`
		RecommendationID   string         `json:"recommendation_id"`
		Status             string         `json:"status"`
		Currency           string         `json:"currency"`
		RollbackReason     string         `json:"rollback_reason"`
		ControlExecution   map[string]any `json:"control_execution"`
		CandidateExecution map[string]any `json:"candidate_execution"`
		Results            map[string]any `json:"results"`
		TrafficPercentage  *string        `json:"traffic_percentage"`
		VerifiedSavings    *string        `json:"verified_savings"`
		Guardrails         Guardrails     `json:"guardrails"`
		StartedAt          *time.Time     `json:"started_at"`
		CompletedAt        *time.Time     `json:"completed_at"`
		Version            int            `json:"version"`
		CreatedAt          time.Time      `json:"created_at"`
		UpdatedAt          time.Time      `json:"updated_at"`
	}
	return json.Marshal(output{e.ID, e.WorkloadID, e.RecommendationID, e.Status, e.Currency, e.RollbackReason, e.ControlExecution, e.CandidateExecution, e.Results, e.TrafficPercentage, e.VerifiedSavings, e.Guardrails, e.StartedAt, e.CompletedAt, e.Version, e.CreatedAt, e.UpdatedAt})
}

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Create(ctx context.Context, orgID, actorID string, input CreateInput) (Experiment, error) {
	if input.WorkloadID == "" || input.RecommendationID == "" || len(input.Currency) != 3 || input.ControlExecution == nil || input.CandidateExecution == nil {
		return Experiment{}, ErrInvalidInput
	}
	if _, err := Violations(input.Guardrails, Metrics{QualityScore: "1", AverageLatencyMS: "0", ErrorRate: "0", CostPerCall: "0"}); err != nil {
		return Experiment{}, err
	}
	control, _ := json.Marshal(input.ControlExecution)
	candidate, _ := json.Marshal(input.CandidateExecution)
	guardrails, _ := json.Marshal(input.Guardrails)
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO experiments(workload_id,recommendation_id,control_execution,candidate_execution,traffic_percentage,guardrails,currency) SELECT w.id,r.id,$4,$5,$6,$7,$8 FROM workloads w JOIN projects p ON p.id=w.project_id JOIN recommendations r ON r.workload_id=w.id WHERE w.id=$1 AND r.id=$2 AND p.organization_id=$3 AND r.status IN('accepted','testing') RETURNING id`, input.WorkloadID, input.RecommendationID, orgID, control, candidate, input.TrafficPercentage, guardrails, strings.ToUpper(input.Currency)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Experiment{}, ErrNotFound
	}
	if err != nil {
		return Experiment{}, err
	}
	s.audit(ctx, orgID, actorID, id, "experiment.created", map[string]any{"status": "draft"})
	return s.Get(ctx, orgID, id)
}

func (s *Store) List(ctx context.Context, orgID string) ([]Experiment, error) {
	rows, err := s.pool.Query(ctx, baseSelect+` WHERE p.organization_id=$1 ORDER BY e.created_at DESC LIMIT 100`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Experiment{}
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Get(ctx context.Context, orgID, id string) (Experiment, error) {
	item, err := scan(s.pool.QueryRow(ctx, baseSelect+` WHERE e.id=$1 AND p.organization_id=$2`, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Experiment{}, ErrNotFound
	}
	return item, err
}

func (s *Store) Transition(ctx context.Context, orgID, actorID, id, action, reason string) (Experiment, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Experiment{}, err
	}
	defer tx.Rollback(ctx)
	var current string
	err = tx.QueryRow(ctx, `SELECT e.status FROM experiments e JOIN workloads w ON w.id=e.workload_id JOIN projects p ON p.id=w.project_id WHERE e.id=$1 AND p.organization_id=$2 FOR UPDATE OF e`, id, orgID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return Experiment{}, ErrNotFound
	}
	if err != nil {
		return Experiment{}, err
	}
	target, err := NextStatus(current, action)
	if err != nil {
		return Experiment{}, err
	}
	reason = strings.TrimSpace(reason)
	if action == "rollback" && reason == "" {
		return Experiment{}, ErrInvalidInput
	}
	_, err = tx.Exec(ctx, `UPDATE experiments SET status=$2,rollback_reason=CASE WHEN $2='rolled_back' THEN $3 ELSE rollback_reason END,started_at=CASE WHEN $2='running' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,completed_at=CASE WHEN $2 IN('completed','verified','rolled_back') THEN CURRENT_TIMESTAMP ELSE completed_at END,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, target, reason)
	if err != nil {
		return Experiment{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Experiment{}, err
	}
	s.audit(ctx, orgID, actorID, id, "experiment."+target, map[string]any{"from_status": current, "reason": reason})
	return s.Get(ctx, orgID, id)
}

func (s *Store) RecordMetrics(ctx context.Context, orgID, actorID, id string, metrics Metrics) (Experiment, error) {
	item, err := s.Get(ctx, orgID, id)
	if err != nil {
		return Experiment{}, err
	}
	if item.Status != "running" {
		return Experiment{}, ErrInvalidTransition
	}
	violations, err := Violations(item.Guardrails, metrics)
	if err != nil {
		return Experiment{}, err
	}
	encoded, _ := json.Marshal(map[string]any{"metrics": metrics, "guardrail_violations": violations})
	if len(violations) > 0 {
		command, execErr := s.pool.Exec(ctx, `UPDATE experiments SET results=$3,status='rolled_back',rollback_reason='automatic guardrail violation',completed_at=CURRENT_TIMESTAMP,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND status='running' AND version=$2`, id, item.Version, encoded)
		if execErr != nil {
			return Experiment{}, execErr
		}
		if command.RowsAffected() != 1 {
			return Experiment{}, ErrInvalidTransition
		}
		s.audit(ctx, orgID, actorID, id, "experiment.auto_rolled_back", map[string]any{"violations": violations})
		return s.Get(ctx, orgID, id)
	}
	command, err := s.pool.Exec(ctx, `UPDATE experiments SET results=$3,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND status='running' AND version=$2`, id, item.Version, encoded)
	if err != nil {
		return Experiment{}, err
	}
	if command.RowsAffected() != 1 {
		return Experiment{}, ErrInvalidTransition
	}
	return s.Get(ctx, orgID, id)
}

func (s *Store) Verify(ctx context.Context, orgID, actorID, id string) (Experiment, error) {
	item, err := s.Get(ctx, orgID, id)
	if err != nil {
		return Experiment{}, err
	}
	if item.Status != "completed" {
		return Experiment{}, ErrInvalidTransition
	}
	raw, ok := item.Results["metrics"]
	if !ok {
		return Experiment{}, ErrInvalidInput
	}
	data, _ := json.Marshal(raw)
	var metrics Metrics
	if json.Unmarshal(data, &metrics) != nil {
		return Experiment{}, ErrInvalidInput
	}
	savings, err := VerifiedSavings(metrics)
	if err != nil || !savings.IsPositive() {
		return Experiment{}, ErrInvalidInput
	}
	command, err := s.pool.Exec(ctx, `UPDATE experiments SET status='verified',verified_savings=$3,completed_at=CURRENT_TIMESTAMP,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND status='completed' AND version=$2`, id, item.Version, savings.StringFixed(12))
	if err != nil {
		return Experiment{}, err
	}
	if command.RowsAffected() != 1 {
		return Experiment{}, ErrInvalidTransition
	}
	s.audit(ctx, orgID, actorID, id, "experiment.verified", map[string]any{"verified_savings": savings.StringFixed(12)})
	return s.Get(ctx, orgID, id)
}

func (s *Store) audit(ctx context.Context, orgID, actorID, id, action string, metadata map[string]any) {
	encoded, _ := json.Marshal(metadata)
	_, _ = s.pool.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id,metadata) VALUES($1,$2,$3,'experiment',$4,$5)`, orgID, actorID, action, id, encoded)
}

const baseSelect = `SELECT e.id,e.workload_id,e.recommendation_id,e.control_execution,e.candidate_execution,e.traffic_percentage::text,e.status,e.guardrails,e.started_at,e.completed_at,e.rollback_reason,e.results,e.verified_savings::text,e.currency,e.version,e.created_at,e.updated_at FROM experiments e JOIN workloads w ON w.id=e.workload_id JOIN projects p ON p.id=w.project_id`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (Experiment, error) {
	var item Experiment
	err := row.Scan(&item.ID, &item.WorkloadID, &item.RecommendationID, &item.ControlExecution, &item.CandidateExecution, &item.TrafficPercentage, &item.Status, &item.Guardrails, &item.StartedAt, &item.CompletedAt, &item.RollbackReason, &item.Results, &item.VerifiedSavings, &item.Currency, &item.Version, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
