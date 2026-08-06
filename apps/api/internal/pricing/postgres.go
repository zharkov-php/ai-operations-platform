package pricing

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) ListModels(ctx context.Context) ([]Model, error) {
	rows, err := s.pool.Query(ctx, `SELECT provider,model,context_window,supports_structured_output,supports_tool_calling,supports_caching,deployment_type,status FROM provider_models ORDER BY provider,model`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Model{}
	for rows.Next() {
		var m Model
		if err = rows.Scan(&m.Provider, &m.Model, &m.ContextWindow, &m.SupportsStructuredOutput, &m.SupportsToolCalling, &m.SupportsCaching, &m.DeploymentType, &m.Status); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}
func (s *Store) CreateModel(ctx context.Context, m Model) (Model, error) {
	if err := ValidateModel(m); err != nil {
		return Model{}, err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO provider_models(provider,model,context_window,supports_structured_output,supports_tool_calling,supports_caching,deployment_type,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, m.Provider, m.Model, m.ContextWindow, m.SupportsStructuredOutput, m.SupportsToolCalling, m.SupportsCaching, m.DeploymentType, m.Status)
	if err != nil {
		return Model{}, mapError(err)
	}
	return m, nil
}
func (s *Store) CreatePrice(ctx context.Context, in PriceInput) (Price, error) {
	var id string
	var cached any
	if in.CachedInputPrice != nil {
		cached = *in.CachedInputPrice
	}
	err := s.pool.QueryRow(ctx, `INSERT INTO model_pricing(provider,model,input_price,output_price,cached_input_price,currency,unit_size,effective_from,effective_to,source_note,illustrative) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, in.Provider, in.Model, in.InputPrice, in.OutputPrice, cached, strings.ToUpper(in.Currency), in.UnitSize, in.EffectiveFrom, in.EffectiveTo, in.SourceNote, in.Illustrative).Scan(&id)
	if err != nil {
		return Price{}, mapError(err)
	}
	return s.GetPrice(ctx, id)
}
func (s *Store) GetPrice(ctx context.Context, id string) (Price, error) {
	return scanPrice(s.pool.QueryRow(ctx, `SELECT id,provider,model,input_price::text,output_price::text,cached_input_price::text,currency,unit_size,effective_from,effective_to,source_note,illustrative,created_at FROM model_pricing WHERE id=$1`, id))
}
func (s *Store) EffectivePrice(ctx context.Context, provider, model, currency string, at time.Time) (Price, error) {
	return scanPrice(s.pool.QueryRow(ctx, `SELECT id,provider,model,input_price::text,output_price::text,cached_input_price::text,currency,unit_size,effective_from,effective_to,source_note,illustrative,created_at FROM model_pricing WHERE provider=$1 AND model=$2 AND currency=$3 AND effective_from<=$4 AND (effective_to IS NULL OR effective_to>$4) ORDER BY effective_from DESC LIMIT 1`, provider, model, strings.ToUpper(currency), at))
}
func (s *Store) ListPrices(ctx context.Context, provider, model, currency string, limit, offset int) ([]Price, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.pool.Query(ctx, `SELECT id,provider,model,input_price::text,output_price::text,cached_input_price::text,currency,unit_size,effective_from,effective_to,source_note,illustrative,created_at FROM model_pricing WHERE ($1='' OR provider=$1) AND ($2='' OR model=$2) AND ($3='' OR currency=$3) ORDER BY effective_from DESC,id LIMIT $4 OFFSET $5`, provider, model, strings.ToUpper(currency), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Price{}
	for rows.Next() {
		item, scanErr := scanPrice(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) UpdatePrice(ctx context.Context, id string, in PriceInput) (Price, error) {
	var cached any
	if in.CachedInputPrice != nil {
		cached = *in.CachedInputPrice
	}
	command, err := s.pool.Exec(ctx, `UPDATE model_pricing SET provider=$2,model=$3,input_price=$4,output_price=$5,cached_input_price=$6,currency=$7,unit_size=$8,effective_from=$9,effective_to=$10,source_note=$11,illustrative=$12 WHERE id=$1`, id, in.Provider, in.Model, in.InputPrice, in.OutputPrice, cached, strings.ToUpper(in.Currency), in.UnitSize, in.EffectiveFrom, in.EffectiveTo, in.SourceNote, in.Illustrative)
	if err != nil {
		return Price{}, mapError(err)
	}
	if command.RowsAffected() == 0 {
		return Price{}, ErrNotFound
	}
	return s.GetPrice(ctx, id)
}

type rowScanner interface{ Scan(...any) error }

func scanPrice(row rowScanner) (Price, error) {
	var p Price
	err := row.Scan(&p.ID, &p.Provider, &p.Model, &p.InputPrice, &p.OutputPrice, &p.CachedInputPrice, &p.Currency, &p.UnitSize, &p.EffectiveFrom, &p.EffectiveTo, &p.SourceNote, &p.Illustrative, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Price{}, ErrNotFound
	}
	return p, err
}
func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23P01" || pgErr.Code == "23505") {
		return ErrConflict
	}
	return err
}
