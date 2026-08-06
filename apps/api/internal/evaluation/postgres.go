package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/redaction"
)

type Dataset struct {
	ID                    string    `json:"id"`
	WorkloadID            string    `json:"workload_id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Source                string    `json:"source"`
	PrivacyClassification string    `json:"privacy_classification"`
	CaseCount             int       `json:"case_count"`
	Cases                 []Case    `json:"cases,omitempty"`
	Runs                  []Run     `json:"runs,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateDatasetInput struct {
	WorkloadID            string `json:"workload_id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Source                string `json:"source"`
	PrivacyClassification string `json:"privacy_classification"`
	Cases                 []Case `json:"cases"`
}

type Run struct {
	ID                 string         `json:"id"`
	DatasetID          string         `json:"dataset_id"`
	CandidateExecution map[string]any `json:"candidate_execution"`
	Status             string         `json:"status"`
	StartedAt          *time.Time     `json:"started_at"`
	CompletedAt        *time.Time     `json:"completed_at"`
	TotalCases         int            `json:"total_cases"`
	PassedCases        int            `json:"passed_cases"`
	FailedCases        int            `json:"failed_cases"`
	AverageLatencyMS   string         `json:"average_latency_ms"`
	EstimatedCost      string         `json:"estimated_cost"`
	Currency           string         `json:"currency"`
	Results            map[string]any `json:"results"`
	CreatedAt          time.Time      `json:"created_at"`
}

type Store struct {
	pool     *pgxpool.Pool
	adapters map[string]CandidateAdapter
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, adapters: map[string]CandidateAdapter{"deterministic_mock": DeterministicAdapter{}}}
}

func (s *Store) CreateDataset(ctx context.Context, orgID, actorID string, input CreateDatasetInput) (Dataset, error) {
	input.Name, input.Description, input.Source = strings.TrimSpace(input.Name), strings.TrimSpace(input.Description), strings.TrimSpace(input.Source)
	if input.WorkloadID == "" || input.Name == "" || len(input.Name) > 200 || input.Source == "" || len(input.Source) > 100 || len(input.Cases) == 0 || !allowedPrivacy(input.PrivacyClassification) {
		return Dataset{}, ErrInvalidInput
	}
	for _, item := range input.Cases {
		if ValidateCase(item) != nil {
			return Dataset{}, ErrInvalidInput
		}
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Dataset{}, err
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO evaluation_datasets(workload_id,name,description,source,privacy_classification,case_count) SELECT w.id,$3,$4,$5,$6,$7 FROM workloads w JOIN projects p ON p.id=w.project_id WHERE w.id=$1 AND p.organization_id=$2 RETURNING id`, input.WorkloadID, orgID, input.Name, input.Description, input.Source, input.PrivacyClassification, len(input.Cases)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Dataset{}, ErrNotFound
	}
	if err != nil {
		return Dataset{}, err
	}
	for _, item := range input.Cases {
		item.SanitizedInput = sanitize(item.SanitizedInput)
		item.ExpectedOutput = sanitize(item.ExpectedOutput)
		item.Metadata = redaction.Metadata(item.Metadata, true)
		sanitized, _ := json.Marshal(item.SanitizedInput)
		expected, _ := json.Marshal(item.ExpectedOutput)
		rules, _ := json.Marshal(item.ValidationRules)
		metadata, _ := json.Marshal(item.Metadata)
		if _, err = tx.Exec(ctx, `INSERT INTO evaluation_cases(dataset_id,sanitized_input,expected_output,validation_rules,metadata) VALUES($1,$2,$3,$4,$5)`, id, sanitized, expected, rules, metadata); err != nil {
			return Dataset{}, err
		}
	}
	metadata, _ := json.Marshal(map[string]any{"case_count": len(input.Cases), "privacy_classification": input.PrivacyClassification})
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id,metadata) VALUES($1,$2,'evaluation_dataset.created','evaluation_dataset',$3,$4)`, orgID, actorID, id, metadata); err != nil {
		return Dataset{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Dataset{}, err
	}
	return s.GetDataset(ctx, orgID, id)
}

func (s *Store) ListDatasets(ctx context.Context, orgID string) ([]Dataset, error) {
	rows, err := s.pool.Query(ctx, datasetSelect+` WHERE p.organization_id=$1 ORDER BY d.created_at DESC LIMIT 100`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Dataset{}
	for rows.Next() {
		item, scanErr := scanDataset(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetDataset(ctx context.Context, orgID, id string) (Dataset, error) {
	item, err := scanDataset(s.pool.QueryRow(ctx, datasetSelect+` WHERE d.id=$1 AND p.organization_id=$2`, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Dataset{}, ErrNotFound
	}
	if err != nil {
		return Dataset{}, err
	}
	item.Cases, err = s.cases(ctx, id)
	if err != nil {
		return Dataset{}, err
	}
	item.Runs, err = s.runs(ctx, id)
	return item, err
}

func (s *Store) CreateRun(ctx context.Context, orgID, actorID, datasetID string, config map[string]any) (Run, error) {
	dataset, err := s.GetDataset(ctx, orgID, datasetID)
	if err != nil {
		return Run{}, err
	}
	adapterName, _ := config["adapter"].(string)
	adapter := s.adapters[adapterName]
	if adapter == nil {
		return Run{}, ErrInvalidInput
	}
	currency, _ := config["currency"].(string)
	currency = strings.ToUpper(currency)
	if len(currency) != 3 {
		return Run{}, ErrInvalidInput
	}
	encoded, _ := json.Marshal(config)
	var id string
	err = s.pool.QueryRow(ctx, `INSERT INTO evaluation_runs(dataset_id,candidate_execution,total_cases,currency) VALUES($1,$2,$3,$4) RETURNING id`, datasetID, encoded, len(dataset.Cases), currency).Scan(&id)
	if err != nil {
		return Run{}, err
	}
	if err = s.transition(ctx, id, "pending", "running"); err != nil {
		return Run{}, err
	}
	caseResults := make([]CaseResult, 0, len(dataset.Cases))
	totalCost, totalLatency, passed := decimal.Zero, decimal.Zero, 0
	for _, item := range dataset.Cases {
		candidate, executeErr := adapter.Execute(ctx, item, config)
		failures := []string{}
		if executeErr != nil {
			failures = append(failures, "candidate_execution")
		} else {
			failures = ValidateRules(candidate.Output, item.ExpectedOutput, item.ValidationRules)
			totalCost = totalCost.Add(candidate.Cost)
			totalLatency = totalLatency.Add(decimal.NewFromInt(candidate.Latency.Milliseconds()))
		}
		if len(failures) == 0 {
			passed++
		}
		caseResults = append(caseResults, CaseResult{CaseID: item.ID, Passed: len(failures) == 0, Output: candidate.Output, Failures: failures})
	}
	averageLatency := decimal.Zero
	if len(dataset.Cases) > 0 {
		averageLatency = totalLatency.Div(decimal.NewFromInt(int64(len(dataset.Cases))))
	}
	baselineLatency, baselineCost := decimal.Zero, decimal.Zero
	if value, ok := config["baseline_latency_ms"]; ok {
		baselineLatency, _ = number(value)
	}
	if value, ok := config["baseline_cost_per_case"]; ok {
		baselineCost, _ = number(value)
	}
	results := map[string]any{
		"cases": caseResults,
		"score": decimal.NewFromInt(int64(passed)).Div(decimal.NewFromInt(int64(len(dataset.Cases)))).StringFixed(4),
		"comparison": map[string]any{
			"control_average_latency_ms":   baselineLatency.StringFixed(6),
			"candidate_average_latency_ms": averageLatency.StringFixed(6),
			"latency_delta_ms":             averageLatency.Sub(baselineLatency).StringFixed(6),
			"control_estimated_cost":       baselineCost.Mul(decimal.NewFromInt(int64(len(dataset.Cases)))).StringFixed(12),
			"candidate_estimated_cost":     totalCost.StringFixed(12),
			"cost_delta":                   totalCost.Sub(baselineCost.Mul(decimal.NewFromInt(int64(len(dataset.Cases))))).StringFixed(12),
		},
	}
	resultJSON, _ := json.Marshal(results)
	_, err = s.pool.Exec(ctx, `UPDATE evaluation_runs SET passed_cases=$2,failed_cases=$3,average_latency_ms=$4,estimated_cost=$5,results=$6 WHERE id=$1`, id, passed, len(dataset.Cases)-passed, averageLatency.StringFixed(6), totalCost.StringFixed(12), resultJSON)
	if err != nil {
		return Run{}, err
	}
	if err = s.transition(ctx, id, "running", "completed"); err != nil {
		return Run{}, err
	}
	metadata, _ := json.Marshal(map[string]any{"dataset_id": datasetID, "passed_cases": passed, "failed_cases": len(dataset.Cases) - passed})
	_, _ = s.pool.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id,metadata) VALUES($1,$2,'evaluation_run.completed','evaluation_run',$3,$4)`, orgID, actorID, id, metadata)
	return s.GetRun(ctx, orgID, id)
}

func (s *Store) GetRun(ctx context.Context, orgID, id string) (Run, error) {
	item, err := scanRun(s.pool.QueryRow(ctx, runSelect+` JOIN evaluation_datasets d ON d.id=r.dataset_id JOIN workloads w ON w.id=d.workload_id JOIN projects p ON p.id=w.project_id WHERE r.id=$1 AND p.organization_id=$2`, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ErrNotFound
	}
	return item, err
}

func (s *Store) transition(ctx context.Context, id, current, target string) error {
	if err := NextStatus(current, target); err != nil {
		return err
	}
	command, err := s.pool.Exec(ctx, `UPDATE evaluation_runs SET status=$3,started_at=CASE WHEN $3='running' THEN CURRENT_TIMESTAMP ELSE started_at END,completed_at=CASE WHEN $3 IN('completed','failed') THEN CURRENT_TIMESTAMP ELSE completed_at END WHERE id=$1 AND status=$2`, id, current, target)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return ErrInvalidTransition
	}
	return nil
}

func (s *Store) cases(ctx context.Context, datasetID string) ([]Case, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,sanitized_input,expected_output,validation_rules,metadata FROM evaluation_cases WHERE dataset_id=$1 ORDER BY created_at,id`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Case{}
	for rows.Next() {
		var item Case
		if err = rows.Scan(&item.ID, &item.SanitizedInput, &item.ExpectedOutput, &item.ValidationRules, &item.Metadata); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) runs(ctx context.Context, datasetID string) ([]Run, error) {
	rows, err := s.pool.Query(ctx, runSelect+` WHERE r.dataset_id=$1 ORDER BY r.created_at DESC LIMIT 50`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Run{}
	for rows.Next() {
		item, scanErr := scanRun(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func allowedPrivacy(value string) bool {
	return value == "public" || value == "internal" || value == "confidential" || value == "restricted"
}

func sanitize(value any) any {
	switch typed := normalize(value).(type) {
	case string:
		return redaction.Text(typed, true)
	case map[string]any:
		return redaction.Metadata(typed, true)
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = sanitize(item)
		}
		return items
	default:
		return typed
	}
}

const datasetSelect = `SELECT d.id,d.workload_id,d.name,d.description,d.source,d.privacy_classification,d.case_count,d.created_at,d.updated_at FROM evaluation_datasets d JOIN workloads w ON w.id=d.workload_id JOIN projects p ON p.id=w.project_id`
const runSelect = `SELECT r.id,r.dataset_id,r.candidate_execution,r.status,r.started_at,r.completed_at,r.total_cases,r.passed_cases,r.failed_cases,r.average_latency_ms::text,r.estimated_cost::text,r.currency,r.results,r.created_at FROM evaluation_runs r`

type scanner interface{ Scan(...any) error }

func scanDataset(row scanner) (Dataset, error) {
	var item Dataset
	err := row.Scan(&item.ID, &item.WorkloadID, &item.Name, &item.Description, &item.Source, &item.PrivacyClassification, &item.CaseCount, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanRun(row scanner) (Run, error) {
	var item Run
	err := row.Scan(&item.ID, &item.DatasetID, &item.CandidateExecution, &item.Status, &item.StartedAt, &item.CompletedAt, &item.TotalCases, &item.PassedCases, &item.FailedCases, &item.AverageLatencyMS, &item.EstimatedCost, &item.Currency, &item.Results, &item.CreatedAt)
	return item, err
}

func (r Run) String() string { return fmt.Sprintf("%s:%s", r.ID, r.Status) }
