package portfolio

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresTenantIsolationAndSlugConflict(t *testing.T) {
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
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var orgA, orgB, userID string
	if err = tx.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Integration A','integration-a-'||gen_random_uuid()) RETURNING id`).Scan(&orgA); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Integration B','integration-b-'||gen_random_uuid()) RETURNING id`).Scan(&orgB); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,'integration-'||gen_random_uuid()||'@example.test','Integration','unused') RETURNING id`, orgA).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO memberships(organization_id,user_id,role) VALUES($1,$2,'owner')`, orgA, userID); err != nil {
		t.Fatal(err)
	}
	// Commit setup because Store intentionally owns its transactions.
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_entries WHERE organization_id=$1`, orgA)
		_, _ = pool.Exec(ctx, `DELETE FROM workloads WHERE project_id IN(SELECT id FROM projects WHERE organization_id=$1)`, orgA)
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE organization_id=$1`, orgA)
		_, _ = pool.Exec(ctx, `DELETE FROM memberships WHERE organization_id=$1`, orgA)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE organization_id=$1`, orgA)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, orgA, orgB)
	}()
	store := NewStore(pool)
	input := ProjectInput{Name: "Integration", Slug: "integration", Environment: "development", MonthlyBudget: "10.25", Currency: "USD"}
	created, err := store.CreateProject(ctx, orgA, userID, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetProject(ctx, orgB, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant get error=%v", err)
	}
	if _, err = store.CreateProject(ctx, orgA, userID, input); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error=%v", err)
	}
}
