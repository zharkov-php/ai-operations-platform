package evaluation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresEvaluationLifecycleAndTenantIsolation(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	suffix := fmt.Sprint(time.Now().UnixNano())
	var orgID, otherOrg, actorID, projectID, workloadID string
	for query, args := range map[string][]any{
		`INSERT INTO organizations(name,slug) VALUES('Evaluation',$1) RETURNING id`: {"evaluation-" + suffix, &orgID},
		`INSERT INTO organizations(name,slug) VALUES('Other',$1) RETURNING id`:      {"evaluation-other-" + suffix, &otherOrg},
	} {
		if err = pool.QueryRow(ctx, query, args[0]).Scan(args[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Evaluator','test') RETURNING id`, orgID, "evaluator-"+suffix+"@example.test").Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Eval','eval','development',100,'USD') RETURNING id`, orgID).Scan(&projectID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'Classifier','classifier','feature','QA','low','standard','internal') RETURNING id`, projectID).Scan(&workloadID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_entries WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM evaluation_runs WHERE dataset_id IN(SELECT id FROM evaluation_datasets WHERE workload_id=$1)`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM evaluation_datasets WHERE workload_id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, actorID)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, orgID, otherOrg)
	}()
	store := NewStore(pool)
	dataset, err := store.CreateDataset(ctx, orgID, actorID, CreateDatasetInput{
		WorkloadID: workloadID, Name: "Ticket ground truth", Source: "sanitized production samples", PrivacyClassification: "internal",
		Cases: []Case{
			{SanitizedInput: map[string]any{"ticket": "invoice"}, ExpectedOutput: map[string]any{"label": "billing"}, ValidationRules: []Rule{{Type: "classification_match", Field: "label"}}},
			{SanitizedInput: map[string]any{"ticket": "password"}, ExpectedOutput: map[string]any{"label": "access"}, ValidationRules: []Rule{{Type: "classification_match", Field: "label"}}},
		},
	})
	if err != nil || dataset.CaseCount != 2 || len(dataset.Cases) != 2 {
		t.Fatalf("dataset=%+v err=%v", dataset, err)
	}
	if _, err = store.GetDataset(ctx, otherOrg, dataset.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross tenant dataset err=%v", err)
	}
	run, err := store.CreateRun(ctx, orgID, actorID, dataset.ID, map[string]any{
		"adapter": "deterministic_mock", "mode": "fixed", "output": map[string]any{"label": "billing"},
		"latency_ms": float64(10), "cost_per_case": "0.001", "baseline_latency_ms": float64(50), "baseline_cost_per_case": "0.02", "currency": "USD",
	})
	if err != nil || run.Status != "completed" || run.PassedCases != 1 || run.FailedCases != 1 || run.EstimatedCost != "0.002000000000" {
		t.Fatalf("run=%+v err=%v", run, err)
	}
	comparison := run.Results["comparison"].(map[string]any)
	if comparison["cost_delta"] != "-0.038000000000" || run.Results["score"] != "0.5000" {
		t.Fatalf("results=%+v", run.Results)
	}
	if _, err = store.GetRun(ctx, otherOrg, run.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross tenant run err=%v", err)
	}
}
