package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/budget"
	"os"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "budget evaluation failed")
		os.Exit(1)
	}
	defer pool.Close()
	created, suppressed, err := budget.NewStore(pool).EvaluateAll(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "budget evaluation failed")
		os.Exit(1)
	}
	fmt.Println(budget.DedupeSummary(created, suppressed))
}
