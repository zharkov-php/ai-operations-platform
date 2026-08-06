package ingestion

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/pricing"
)

func TestPostgresIngestionRedactionCostAndConcurrency(t *testing.T) {
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
	var orgID, projectID, workloadID string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Ingestion Integration',$1) RETURNING id`, "ingestion-"+suffix).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM llm_calls WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE project_id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, projectID)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, orgID)
	}()
	if err = pool.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES($1,'Ingestion','ingestion','development',10,'USD') RETURNING id`, orgID).Scan(&projectID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification) VALUES($1,'Classifier','classifier','feature','Test','low','standard','internal') RETURNING id`, projectID).Scan(&workloadID); err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool, pricing.NewStore(pool))
	input := Input{ProjectID: projectID, WorkloadID: workloadID, ExternalCallID: "concurrent", Provider: "fictional-cloud", Model: "illustrative-frontier", RequestTimestamp: time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC), LatencyMS: 10, InputTokens: 1000, OutputTokens: 200, CachedInputTokens: 500, Status: "success", Prompt: "Bearer super-secret owner@example.test", Response: "password=hunter2", Metadata: map[string]any{"api_key": "hidden"}}
	const workers = 8
	ids := make(chan string, workers)
	errs := make(chan error, workers)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			call, ingestErr := store.Ingest(ctx, orgID, nil, input)
			if ingestErr != nil {
				errs <- ingestErr
				return
			}
			if call.EstimatedCost != "0.008500000000" {
				errs <- fmt.Errorf("cost=%s", call.EstimatedCost)
				return
			}
			ids <- call.ID
		}()
	}
	group.Wait()
	close(ids)
	close(errs)
	for ingestErr := range errs {
		t.Fatal(ingestErr)
	}
	unique := map[string]bool{}
	for id := range ids {
		unique[id] = true
	}
	if len(unique) != 1 {
		t.Fatalf("unique IDs=%d", len(unique))
	}
	var prompt, response, metadata string
	if err = pool.QueryRow(ctx, `SELECT COALESCE(redacted_prompt_preview,''),COALESCE(redacted_response_preview,''),metadata::text FROM llm_calls WHERE organization_id=$1`, orgID).Scan(&prompt, &response, &metadata); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"super-secret", "owner@example.test", "hunter2", "hidden"} {
		if strings.Contains(prompt+response+metadata, secret) {
			t.Fatalf("persisted secret %q", secret)
		}
	}
}
