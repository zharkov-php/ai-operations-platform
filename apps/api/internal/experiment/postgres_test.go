package experiment

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestPostgresExperimentGuardrailsRollbackAndVerification(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	suffix := fmt.Sprint(time.Now().UnixNano())
	var org, actor, project, workload, recommendation string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Experiments',$1) RETURNING id`, "experiments-"+suffix).Scan(&org); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Owner','test') RETURNING id`, org, "experiment-"+suffix+"@example.test").Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Experiment','experiment','production',1000,'USD') RETURNING id`, org).Scan(&project); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'Classifier','classifier','feature','QA','low','standard','internal') RETURNING id`, project).Scan(&workload); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO recommendations(workload_id,recommendation_type,priority,confidence_level,confidence,status,current_execution,proposed_execution,reason_codes,evidence_summary,confidence_inputs,currency,quality_risk,operational_risk,required_next_action,rule_version) VALUES($1,'smaller_hosted_model','high','high',0.9,'accepted','{}','{}','{}','{}','{}','USD','medium','low','evaluation_required','test') RETURNING id`, workload).Scan(&recommendation); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_entries WHERE organization_id=$1`, org)
		_, _ = pool.Exec(ctx, `DELETE FROM experiments WHERE workload_id=$1`, workload)
		_, _ = pool.Exec(ctx, `DELETE FROM recommendations WHERE id=$1`, recommendation)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE id=$1`, workload)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, project)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, actor)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, org)
	}()
	store := NewStore(pool)
	create := func() Experiment {
		item, createErr := store.Create(ctx, org, actor, CreateInput{WorkloadID: workload, RecommendationID: recommendation, TrafficPercentage: "10", Currency: "USD", ControlExecution: map[string]any{"model": "frontier"}, CandidateExecution: map[string]any{"model": "compact"}, Guardrails: Guardrails{"0.90", "500", "0.05", "0.02"}})
		if createErr != nil {
			t.Fatal(createErr)
		}
		item, createErr = store.Transition(ctx, org, actor, item.ID, "approve", "")
		if createErr != nil {
			t.Fatal(createErr)
		}
		item, createErr = store.Transition(ctx, org, actor, item.ID, "start", "")
		if createErr != nil {
			t.Fatal(createErr)
		}
		return item
	}
	failed := create()
	failed, err = store.RecordMetrics(ctx, org, actor, failed.ID, Metrics{QualityScore: "0.7", AverageLatencyMS: "400", ErrorRate: "0.01", CostPerCall: "0.01"})
	if err != nil || failed.Status != "rolled_back" || failed.RollbackReason != "automatic guardrail violation" {
		t.Fatalf("failed=%+v err=%v", failed, err)
	}
	success := create()
	metrics := Metrics{QualityScore: "0.95", AverageLatencyMS: "400", ErrorRate: "0.01", CostPerCall: "0.01", ControlTotalCost: "100", CandidateTotalCost: "40"}
	success, err = store.RecordMetrics(ctx, org, actor, success.ID, metrics)
	if err != nil {
		t.Fatal(err)
	}
	success, err = store.Transition(ctx, org, actor, success.ID, "complete", "")
	if err != nil {
		t.Fatal(err)
	}
	success, err = store.Verify(ctx, org, actor, success.ID)
	if err != nil || success.Status != "verified" || *success.VerifiedSavings != "60.000000000000" {
		t.Fatalf("success=%+v err=%v", success, err)
	}
	concurrent := create()
	results := make(chan error, 2)
	go func() {
		_, actionErr := store.Transition(ctx, org, actor, concurrent.ID, "pause", "")
		results <- actionErr
	}()
	go func() {
		_, actionErr := store.Transition(ctx, org, actor, concurrent.ID, "pause", "")
		results <- actionErr
	}()
	first, second := <-results, <-results
	if (first == nil) == (second == nil) {
		t.Fatalf("concurrent results first=%v second=%v", first, second)
	}
}
