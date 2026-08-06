package pricing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresEffectivePricingAndOverlap(t *testing.T) {
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
	store := NewStore(pool)
	provider := "integration"
	model := fmt.Sprintf("model-%d", time.Now().UnixNano())
	_, err = store.CreateModel(ctx, Model{Provider: provider, Model: model, ContextWindow: 8192, DeploymentType: "hosted", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM provider_models WHERE provider=$1 AND model=$2`, provider, model)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	input := PriceInput{Provider: provider, Model: model, InputPrice: "0.123456789012", OutputPrice: "2", Currency: "USD", UnitSize: 1000000, EffectiveFrom: from, EffectiveTo: &to, SourceNote: "Integration illustrative", Illustrative: true}
	created, err := store.CreatePrice(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM model_pricing WHERE provider=$1 AND model=$2`, provider, model)
	if created.InputPrice != "0.123456789012" {
		t.Fatalf("precision=%s", created.InputPrice)
	}
	found, err := store.EffectivePrice(ctx, provider, model, "usd", from)
	if err != nil || found.ID != created.ID {
		t.Fatalf("found=%+v err=%v", found, err)
	}
	if _, err = store.EffectivePrice(ctx, provider, model, "USD", to); !errors.Is(err, ErrNotFound) {
		t.Fatalf("exclusive boundary error=%v", err)
	}
	overlap := input
	overlap.EffectiveFrom = from.Add(time.Hour)
	if _, err = store.CreatePrice(ctx, overlap); !errors.Is(err, ErrConflict) {
		t.Fatalf("overlap error=%v", err)
	}
}
