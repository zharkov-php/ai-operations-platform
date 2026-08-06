package recommendation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresAnalysisPersistsVersionedTenantScopedRecommendations(t *testing.T) {
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
	var orgID, otherOrg, actorID, projectID, workloadID string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Recommendations',$1) RETURNING id`, "recommendations-"+suffix).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Other',$1) RETURNING id`, "recommendations-other-"+suffix).Scan(&otherOrg); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Reviewer','test-hash') RETURNING id`, orgID, "reviewer-"+suffix+"@example.test").Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Rules','rules','production',1000,'USD') RETURNING id`, orgID).Scan(&projectID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'VAT Calculator','vat','feature','Finance','medium','standard','internal') RETURNING id`, projectID).Scan(&workloadID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_entries WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM recommendations WHERE workload_id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM task_classifications WHERE workload_id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM llm_calls WHERE workload_id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE id=$1`, workloadID)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, actorID)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, orgID, otherOrg)
	}()
	_, err = pool.Exec(ctx, `INSERT INTO llm_calls(organization_id,project_id,workload_id,external_call_id,provider,model,prompt_template_id,request_timestamp,latency_ms,input_tokens,output_tokens,cached_input_tokens,retry_count,status,estimated_cost,currency,prompt_hash,metadata,payload_hash) SELECT $1,$2,$3,'rule-'||g,'fictional-cloud','illustrative-frontier','vat-template',CURRENT_TIMESTAMP-interval '1 day',50,200,20,0,0,'success',1,'USD',digest('prompt-'||(g%10),'sha256'),'{"task_type":"calculation","determinism":"high","complexity":"low","failure_impact":"medium"}'::jsonb,digest('payload-'||g,'sha256') FROM generate_series(1,100) g`, orgID, projectID, workloadID)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	first, err := store.AnalyzeOrganization(ctx, orgID)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRecommendation(first, "replace_with_code") || !hasRecommendation(first, "exact_cache") {
		t.Fatalf("recommendations=%+v", first)
	}
	second, err := store.AnalyzeOrganization(ctx, orgID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("first=%d second=%d", len(first), len(second))
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM recommendations WHERE workload_id=$1 AND rule_version=$2`, workloadID, RuleVersion).Scan(&count); err != nil || count != len(first) {
		t.Fatalf("count=%d err=%v", count, err)
	}
	var classifications int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM task_classifications WHERE workload_id=$1`, workloadID).Scan(&classifications)
	if classifications != 1 {
		t.Fatalf("classifications=%d", classifications)
	}
	if _, err = store.Get(ctx, otherOrg, first[0].ID); err != ErrNotFound {
		t.Fatalf("cross tenant error=%v", err)
	}
	accepted, err := store.Review(ctx, orgID, actorID, first[0].ID, "accept", "approved for evaluation")
	if err != nil || accepted.Status != "accepted" {
		t.Fatalf("accepted=%+v err=%v", accepted, err)
	}
	if _, err = store.Review(ctx, orgID, actorID, first[0].ID, "accept", "duplicate"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("duplicate transition err=%v", err)
	}
	if _, err = store.Review(ctx, otherOrg, actorID, first[1].ID, "reject", "not suitable"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross tenant review err=%v", err)
	}
	if _, err = store.Review(ctx, orgID, actorID, first[1].ID, "reject", " "); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("empty reason err=%v", err)
	}
	rejected, err := store.Review(ctx, orgID, actorID, first[1].ID, "reject", "quality risk is too high")
	if err != nil || rejected.Status != "rejected" {
		t.Fatalf("rejected=%+v err=%v", rejected, err)
	}
	history, err := store.AuditHistory(ctx, orgID, first[0].ID)
	if err != nil || len(history) != 1 || history[0].Action != "recommendation.accepted" || history[0].Metadata["effect"] != "authorizes_evaluation_only" {
		t.Fatalf("history=%+v err=%v", history, err)
	}
}
func hasRecommendation(items []Recommendation, kind string) bool {
	for _, item := range items {
		if item.Type == kind {
			return true
		}
	}
	return false
}
