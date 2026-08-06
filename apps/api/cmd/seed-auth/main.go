package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
)

func main() {
	if err := run(context.Background(), os.Getenv("DATABASE_URL"), os.Getenv("SEED_OWNER_PASSWORD")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("development owner seeded: owner@example.test")
}
func run(ctx context.Context, databaseURL, password string) error {
	if databaseURL == "" || password == "" {
		return fmt.Errorf("DATABASE_URL and SEED_OWNER_PASSWORD are required")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var orgID string
	if err = tx.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES('Demo Organization','demo') ON CONFLICT(slug) DO UPDATE SET name=EXCLUDED.name RETURNING id`).Scan(&orgID); err != nil {
		return err
	}
	var userID string
	if err = tx.QueryRow(ctx, `INSERT INTO users(organization_id,email,display_name,password_hash) VALUES($1,'owner@example.test','Demo Owner',$2) ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash,status='active',organization_id=EXCLUDED.organization_id RETURNING id`, orgID, hash).Scan(&userID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO memberships(organization_id,user_id,role) VALUES($1,$2,'owner') ON CONFLICT(organization_id,user_id) DO UPDATE SET role='owner'`, orgID, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
