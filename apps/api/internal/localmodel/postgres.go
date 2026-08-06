package localmodel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoredConfiguration struct {
	Configuration
	OrganizationID string    `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Create(ctx context.Context, orgID, actorID string, input Configuration) (StoredConfiguration, error) {
	input.HardwareName, input.SupportedModel, input.BenchmarkSource = strings.TrimSpace(input.HardwareName), strings.TrimSpace(input.SupportedModel), strings.TrimSpace(input.BenchmarkSource)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if input.HardwareName == "" || len(input.HardwareName) > 200 || input.SupportedModel == "" || input.ContextLimit <= 0 || input.BenchmarkSource == "" {
		return StoredConfiguration{}, ErrInvalidInput
	}
	if _, err := Compare(input, ComparisonInput{HostedCostPerRequest: "0", RequiredMemoryGB: "0"}); err != nil {
		return StoredConfiguration{}, err
	}
	var id string
	err := s.pool.QueryRow(ctx, `INSERT INTO local_model_configurations(organization_id,hardware_name,purchase_cost,useful_lifetime_months,monthly_electricity,monthly_maintenance,available_memory_gb,estimated_requests_per_second,utilization,supported_model,context_limit,currency,benchmark_source) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`, orgID, input.HardwareName, input.PurchaseCost, input.UsefulLifetimeMonths, input.MonthlyElectricity, input.MonthlyMaintenance, input.AvailableMemoryGB, input.EstimatedRequestsPerSecond, input.Utilization, input.SupportedModel, input.ContextLimit, input.Currency, input.BenchmarkSource).Scan(&id)
	if err != nil {
		return StoredConfiguration{}, err
	}
	metadata, _ := json.Marshal(map[string]any{"hardware_name": input.HardwareName, "benchmark_source": input.BenchmarkSource})
	_, _ = s.pool.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id,metadata) VALUES($1,$2,'local_model_configuration.created','local_model_configuration',$3,$4)`, orgID, actorID, id, metadata)
	return s.Get(ctx, orgID, id)
}

func (s *Store) List(ctx context.Context, orgID string) ([]StoredConfiguration, error) {
	rows, err := s.pool.Query(ctx, selectConfiguration+` WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 100`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []StoredConfiguration{}
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) Get(ctx context.Context, orgID, id string) (StoredConfiguration, error) {
	item, err := scan(s.pool.QueryRow(ctx, selectConfiguration+` WHERE id=$1 AND organization_id=$2`, id, orgID))
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredConfiguration{}, ErrNotFound
	}
	return item, err
}

const selectConfiguration = `SELECT id,organization_id,hardware_name,purchase_cost::text,useful_lifetime_months,monthly_electricity::text,monthly_maintenance::text,available_memory_gb::text,estimated_requests_per_second::text,utilization::text,supported_model,context_limit,currency,benchmark_source,created_at,updated_at FROM local_model_configurations`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (StoredConfiguration, error) {
	var item StoredConfiguration
	err := row.Scan(&item.ID, &item.OrganizationID, &item.HardwareName, &item.PurchaseCost, &item.UsefulLifetimeMonths, &item.MonthlyElectricity, &item.MonthlyMaintenance, &item.AvailableMemoryGB, &item.EstimatedRequestsPerSecond, &item.Utilization, &item.SupportedModel, &item.ContextLimit, &item.Currency, &item.BenchmarkSource, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
