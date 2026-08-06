package budget

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"time"
)

type Threshold struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Percentage string    `json:"percentage"`
	Severity   string    `json:"severity"`
	CreatedAt  time.Time `json:"created_at"`
}
type Alert struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	ThresholdType  string         `json:"threshold_type"`
	ThresholdValue string         `json:"threshold_value"`
	Severity       string         `json:"severity"`
	Status         string         `json:"status"`
	DedupeKey      string         `json:"dedupe_key"`
	Evidence       map[string]any `json:"evidence"`
	TriggeredAt    time.Time      `json:"triggered_at"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at"`
	AcknowledgedBy *string        `json:"acknowledged_by"`
}
type Store struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool, now: time.Now} }
func (s *Store) EvaluateAll(ctx context.Context) (int, int, error) {
	rows, err := s.pool.Query(ctx, `SELECT id FROM organizations WHERE status='active' ORDER BY id`)
	if err != nil {
		return 0, 0, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	created, suppressed := 0, 0
	for _, id := range ids {
		newCount, duplicateCount, evaluateErr := s.EvaluateOrganization(ctx, id)
		if evaluateErr != nil {
			return created, suppressed, evaluateErr
		}
		created += newCount
		suppressed += duplicateCount
	}
	return created, suppressed, nil
}
func (s *Store) SetThreshold(ctx context.Context, orgID, projectID, percentage, severity string) (Threshold, error) {
	value, err := ValidateThreshold(percentage)
	if err != nil || (severity != "info" && severity != "warning" && severity != "critical") {
		return Threshold{}, ErrInvalidInput
	}
	var id string
	err = s.pool.QueryRow(ctx, `INSERT INTO project_budget_thresholds(project_id,percentage,severity) SELECT p.id,$3,$4 FROM projects p WHERE p.id=$1 AND p.organization_id=$2 ON CONFLICT(project_id,percentage) DO UPDATE SET severity=EXCLUDED.severity RETURNING id`, projectID, orgID, value.StringFixed(2), severity).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Threshold{}, ErrNotFound
	}
	if err != nil {
		return Threshold{}, err
	}
	return s.getThreshold(ctx, orgID, id)
}
func (s *Store) getThreshold(ctx context.Context, orgID, id string) (Threshold, error) {
	var item Threshold
	err := s.pool.QueryRow(ctx, `SELECT t.id,t.project_id,t.percentage::text,t.severity,t.created_at FROM project_budget_thresholds t JOIN projects p ON p.id=t.project_id WHERE t.id=$1 AND p.organization_id=$2`, id, orgID).Scan(&item.ID, &item.ProjectID, &item.Percentage, &item.Severity, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Threshold{}, ErrNotFound
	}
	return item, err
}
func (s *Store) ListThresholds(ctx context.Context, orgID string) ([]Threshold, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.id,t.project_id,t.percentage::text,t.severity,t.created_at FROM project_budget_thresholds t JOIN projects p ON p.id=t.project_id WHERE p.organization_id=$1 ORDER BY t.percentage`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Threshold{}
	for rows.Next() {
		var item Threshold
		if err = rows.Scan(&item.ID, &item.ProjectID, &item.Percentage, &item.Severity, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) ListAlerts(ctx context.Context, orgID, status string) ([]Alert, error) {
	rows, err := s.pool.Query(ctx, `SELECT a.id,a.project_id,a.threshold_type,a.threshold_value::text,a.severity,a.status,a.dedupe_key,a.evidence,a.triggered_at,a.acknowledged_at,a.acknowledged_by::text FROM budget_alerts a JOIN projects p ON p.id=a.project_id WHERE p.organization_id=$1 AND ($2='' OR a.status=$2) ORDER BY a.triggered_at DESC LIMIT 100`, orgID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Alert{}
	for rows.Next() {
		item, scanErr := scanAlert(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Acknowledge(ctx context.Context, orgID, actorID, id string) (Alert, error) {
	command, err := s.pool.Exec(ctx, `UPDATE budget_alerts a SET status='acknowledged',acknowledged_at=CURRENT_TIMESTAMP,acknowledged_by=$3 FROM projects p WHERE a.project_id=p.id AND a.id=$1 AND p.organization_id=$2 AND a.status='open'`, id, orgID, actorID)
	if err != nil {
		return Alert{}, err
	}
	if command.RowsAffected() != 1 {
		return Alert{}, ErrInvalidTransition
	}
	return s.getAlert(ctx, orgID, id)
}
func (s *Store) getAlert(ctx context.Context, orgID, id string) (Alert, error) {
	item, err := scanAlert(s.pool.QueryRow(ctx, `SELECT a.id,a.project_id,a.threshold_type,a.threshold_value::text,a.severity,a.status,a.dedupe_key,a.evidence,a.triggered_at,a.acknowledged_at,a.acknowledged_by::text FROM budget_alerts a JOIN projects p ON p.id=a.project_id WHERE a.id=$1 AND p.organization_id=$2`, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Alert{}, ErrNotFound
	}
	return item, err
}
func (s *Store) EvaluateOrganization(ctx context.Context, orgID string) (int, int, error) {
	rows, err := s.pool.Query(ctx, `SELECT p.id,p.currency,p.monthly_budget::text,COALESCE((SELECT sum(c.estimated_cost) FROM llm_calls c WHERE c.project_id=p.id AND c.request_timestamp>=date_trunc('month',CURRENT_TIMESTAMP)),0)::text,COALESCE((SELECT sum(c.estimated_cost) FROM llm_calls c WHERE c.project_id=p.id AND c.request_timestamp>=date_trunc('day',CURRENT_TIMESTAMP)),0)::text,COALESCE((SELECT avg(d.cost) FROM(SELECT sum(c2.estimated_cost) cost FROM llm_calls c2 WHERE c2.project_id=p.id AND c2.request_timestamp>=date_trunc('day',CURRENT_TIMESTAMP)-interval '7 days' AND c2.request_timestamp<date_trunc('day',CURRENT_TIMESTAMP) GROUP BY date_trunc('day',c2.request_timestamp))d),0)::text FROM projects p WHERE p.organization_id=$1 AND p.status='active'`, orgID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	snapshots := []Snapshot{}
	for rows.Next() {
		var snapshot Snapshot
		var budget, cost, today, prior string
		if err = rows.Scan(&snapshot.ProjectID, &snapshot.Currency, &budget, &cost, &today, &prior); err != nil {
			return 0, 0, err
		}
		snapshot.MonthlyBudget, _ = decimal.NewFromString(budget)
		snapshot.MonthCost, _ = decimal.NewFromString(cost)
		snapshot.TodayCost, _ = decimal.NewFromString(today)
		snapshot.PriorDailyAverage, _ = decimal.NewFromString(prior)
		snapshot.Now = s.now()
		snapshots = append(snapshots, snapshot)
	}
	created, suppressed := 0, 0
	for _, snapshot := range snapshots {
		thresholdRows, queryErr := s.pool.Query(ctx, `SELECT percentage::text,severity FROM project_budget_thresholds WHERE project_id=$1`, snapshot.ProjectID)
		if queryErr != nil {
			return created, suppressed, queryErr
		}
		thresholds := []ThresholdRule{}
		for thresholdRows.Next() {
			var percentage, severity string
			if queryErr = thresholdRows.Scan(&percentage, &severity); queryErr != nil {
				thresholdRows.Close()
				return created, suppressed, queryErr
			}
			value, _ := decimal.NewFromString(percentage)
			thresholds = append(thresholds, ThresholdRule{Percentage: value, Severity: severity})
		}
		thresholdRows.Close()
		for _, candidate := range Detect(snapshot, thresholds) {
			evidence, _ := json.Marshal(candidate.Evidence)
			command, execErr := s.pool.Exec(ctx, `INSERT INTO budget_alerts(project_id,threshold_type,threshold_value,severity,dedupe_key,evidence) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(project_id,dedupe_key) WHERE status='open' DO NOTHING`, snapshot.ProjectID, candidate.Type, candidate.ThresholdValue.StringFixed(12), candidate.Severity, candidate.DedupeKey, evidence)
			if execErr != nil {
				return created, suppressed, execErr
			}
			if command.RowsAffected() == 1 {
				created++
			} else {
				suppressed++
			}
		}
	}
	return created, suppressed, nil
}

type scanner interface{ Scan(...any) error }

func scanAlert(row scanner) (Alert, error) {
	var item Alert
	err := row.Scan(&item.ID, &item.ProjectID, &item.ThresholdType, &item.ThresholdValue, &item.Severity, &item.Status, &item.DedupeKey, &item.Evidence, &item.TriggeredAt, &item.AcknowledgedAt, &item.AcknowledgedBy)
	return item, err
}
