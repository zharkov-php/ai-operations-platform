package budget

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestPostgresBudgetEvaluationDeduplicationAndAcknowledgement(t *testing.T) {
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
	var org, other, actor, project, workload string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Budget',$1) RETURNING id`, "budget-"+suffix).Scan(&org); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Other',$1) RETURNING id`, "budget-other-"+suffix).Scan(&other); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Finance','test') RETURNING id`, org, "budget-"+suffix+"@example.test").Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Budget','budget','production',1000,'USD') RETURNING id`, org).Scan(&project); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'Usage','usage','feature','Ops','low','standard','internal') RETURNING id`, project).Scan(&workload); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM budget_alerts WHERE project_id=$1`, project)
		_, _ = pool.Exec(ctx, `DELETE FROM project_budget_thresholds WHERE project_id=$1`, project)
		_, _ = pool.Exec(ctx, `DELETE FROM llm_calls WHERE workload_id=$1`, workload)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE id=$1`, workload)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, project)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, actor)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, org, other)
	}()
	_, err = pool.Exec(ctx, `INSERT INTO llm_calls(organization_id,project_id,workload_id,external_call_id,provider,model,request_timestamp,latency_ms,input_tokens,output_tokens,cached_input_tokens,retry_count,status,estimated_cost,currency,prompt_hash,metadata,payload_hash) VALUES($1,$2,$3,$4,'fictional-cloud','illustrative-frontier',CURRENT_TIMESTAMP,10,1,1,0,0,'success',600,'USD',digest('budget','sha256'),'{}',digest('budget-payload','sha256'))`, org, project, workload, "budget-"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	threshold, err := store.SetThreshold(ctx, org, project, "50", "warning")
	if err != nil || threshold.Percentage != "50.00" {
		t.Fatalf("threshold=%+v err=%v", threshold, err)
	}
	created, suppressed, err := store.EvaluateOrganization(ctx, org)
	if err != nil || created != 2 || suppressed != 0 {
		t.Fatalf("first created=%d suppressed=%d err=%v", created, suppressed, err)
	}
	created, suppressed, err = store.EvaluateOrganization(ctx, org)
	if err != nil || created != 0 || suppressed != 2 {
		t.Fatalf("second created=%d suppressed=%d err=%v", created, suppressed, err)
	}
	alerts, err := store.ListAlerts(ctx, org, "open")
	if err != nil || len(alerts) != 2 {
		t.Fatalf("alerts=%+v err=%v", alerts, err)
	}
	if _, err = store.Acknowledge(ctx, other, actor, alerts[0].ID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("cross tenant err=%v", err)
	}
	acknowledged, err := store.Acknowledge(ctx, org, actor, alerts[0].ID)
	if err != nil || acknowledged.Status != "acknowledged" || acknowledged.AcknowledgedAt == nil {
		t.Fatalf("ack=%+v err=%v", acknowledged, err)
	}
	if _, err = store.Acknowledge(ctx, org, actor, alerts[0].ID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("duplicate ack err=%v", err)
	}
}
