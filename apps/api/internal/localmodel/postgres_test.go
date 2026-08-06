package localmodel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresLocalModelConfigurationIsTenantScoped(t *testing.T) {
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
	var orgID, otherOrg, actorID string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Local',$1) RETURNING id`, "local-"+suffix).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Other',$1) RETURNING id`, "local-other-"+suffix).Scan(&otherOrg); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Owner','test') RETURNING id`, orgID, "local-"+suffix+"@example.test").Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_entries WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM local_model_configurations WHERE organization_id=$1`, orgID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, actorID)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id IN($1,$2)`, orgID, otherOrg)
	}()
	store := NewStore(pool)
	created, err := store.Create(ctx, orgID, actorID, testConfig())
	if err != nil || created.OrganizationID != orgID || created.PurchaseCost != "3600.000000000000" {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	if _, err = store.Get(ctx, otherOrg, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross tenant err=%v", err)
	}
	items, err := store.List(ctx, orgID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}
