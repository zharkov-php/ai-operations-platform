package analytics

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresAnalyticsBoundariesAggregationsAndTenantIsolation(t *testing.T) {
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
	var orgID, otherOrg, projectID, workloadID string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Analytics',$1) RETURNING id`, "analytics-"+suffix).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Other',$1) RETURNING id`, "analytics-other-"+suffix).Scan(&otherOrg); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM llm_calls WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE project_id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, orgID, otherOrg)
	}()
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Analytics','analytics','production',10,'USD') RETURNING id`, orgID).Scan(&projectID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'Workload','workload','feature','Test','low','standard','internal') RETURNING id`, projectID).Scan(&workloadID); err != nil {
		t.Fatal(err)
	}
	insert := func(id string, at time.Time, cost string, status string, retries int, task string) {
		t.Helper()
		_, insertErr := pool.Exec(ctx, `INSERT INTO llm_calls(organization_id,project_id,workload_id,external_call_id,provider,model,request_timestamp,latency_ms,input_tokens,output_tokens,cached_input_tokens,retry_count,status,estimated_cost,currency,metadata,payload_hash) VALUES($1,$2,$3,$4,'fictional-cloud','illustrative-frontier',$5,100,10,2,1,$6,$7,$8,'USD',jsonb_build_object('task_type',$9::text),decode(repeat('00',32),'hex'))`, orgID, projectID, workloadID, id, at, retries, status, cost, task)
		if insertErr != nil {
			t.Fatal(insertErr)
		}
	}
	insert("previous", time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC), "0.5", "success", 0, "classification")
	insert("included-a", time.Date(2026, 8, 5, 23, 0, 0, 0, time.UTC), "1", "success", 0, "classification")
	insert("included-b", time.Date(2026, 8, 6, 22, 59, 0, 0, time.UTC), "2", "error", 1, "classification")
	insert("excluded-end", time.Date(2026, 8, 6, 23, 0, 0, 0, time.UTC), "4", "success", 0, "classification")
	rng, err := ParseRange("2026-08-06", "2026-08-07", "Atlantic/Canary", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	filters := Filters{Range: rng, Currency: "USD"}
	store := NewStore(pool)
	overview, err := store.Overview(ctx, orgID, filters)
	if err != nil {
		t.Fatal(err)
	}
	if overview.CallCount != 2 || overview.ObservedCost != "3.000000000000" || overview.PreviousPeriodCost != "0.500000000000" || overview.ErrorRate != "0.50000000000000000000" || overview.RetryRate != "0.50000000000000000000" {
		t.Fatalf("overview=%+v", overview)
	}
	daily, err := store.Daily(ctx, orgID, filters)
	if err != nil || len(daily) != 1 || daily[0].Date != "2026-08-06" {
		t.Fatalf("daily=%+v err=%v", daily, err)
	}
	tasks, err := store.Costs(ctx, orgID, "task", filters)
	if err != nil || len(tasks) != 1 || tasks[0].Key != "classification" {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
	other, err := store.Overview(ctx, otherOrg, filters)
	if err != nil || other.CallCount != 0 {
		t.Fatalf("other=%+v err=%v", other, err)
	}
}
