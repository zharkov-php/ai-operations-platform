package notification

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestPostgresNotificationPreferencesAndDeduplication(t *testing.T) {
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
	var org, user string
	if err = pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Notify',$1) RETURNING id`, "notify-"+suffix).Scan(&org); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,$2,'Owner','test') RETURNING id`, org, "notify-"+suffix+"@example.test").Scan(&user); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, user)
		_, _ = pool.Exec(ctx, `DELETE FROM organizations WHERE id=$1`, org)
	}()
	store := NewStore(pool)
	preference, err := store.SetPreferences(ctx, user, Preferences{InApp: true, MobilePush: false})
	if err != nil || !preference.InApp || preference.MobilePush {
		t.Fatalf("preferences=%+v err=%v", preference, err)
	}
	event := Event{Type: "evaluation_completed", Title: "Evaluation completed", Body: "Results are ready.", DeepLink: DeepLink("evaluations", "one"), DedupeKey: "evaluation:one"}
	_, created, err := store.Publish(ctx, org, user, event)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	_, created, err = store.Publish(ctx, org, user, event)
	if err != nil || created {
		t.Fatalf("duplicate created=%v err=%v", created, err)
	}
	items, err := store.List(ctx, org, user)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}
