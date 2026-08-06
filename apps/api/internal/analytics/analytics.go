package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var ErrInvalidRange = errors.New("invalid analytics range")

type Range struct {
	From, To time.Time
	Timezone string
}
type Filters struct {
	Range                                            Range
	ProjectID, WorkloadID, Provider, Model, Currency string
}
type Overview struct {
	ObservedCost               string `json:"observed_cost"`
	PreviousPeriodCost         string `json:"previous_period_cost"`
	EstimatedMonthlyProjection string `json:"estimated_monthly_projection"`
	Currency                   string `json:"currency"`
	CallCount                  int64  `json:"call_count"`
	InputTokens                int64  `json:"input_tokens"`
	OutputTokens               int64  `json:"output_tokens"`
	CachedInputTokens          int64  `json:"cached_input_tokens"`
	AverageLatencyMS           string `json:"average_latency_ms"`
	ErrorRate                  string `json:"error_rate"`
	RetryRate                  string `json:"retry_rate"`
	ProjectionMethod           string `json:"projection_method"`
}
type CostPoint struct {
	Key      string `json:"key"`
	Cost     string `json:"cost"`
	Calls    int64  `json:"calls"`
	Currency string `json:"currency"`
}
type DailyPoint struct {
	Date  string `json:"date"`
	Cost  string `json:"cost"`
	Calls int64  `json:"calls"`
}
type TokenSummary struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CachedInput int64 `json:"cached_input"`
}
type LatencySummary struct {
	AverageMS string `json:"average_ms"`
	P50MS     string `json:"p50_ms"`
	P95MS     string `json:"p95_ms"`
}
type ErrorSummary struct {
	Calls        int64  `json:"calls"`
	Errors       int64  `json:"errors"`
	RetriedCalls int64  `json:"retried_calls"`
	ErrorRate    string `json:"error_rate"`
	RetryRate    string `json:"retry_rate"`
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func (s *Store) DefaultCurrency(ctx context.Context, orgID string) (string, error) {
	var currency string
	err := s.pool.QueryRow(ctx, `SELECT default_currency FROM organizations WHERE id=$1`, orgID).Scan(&currency)
	return currency, err
}

func ParseRange(from, to, timezone string, now time.Time) (Range, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Range{}, ErrInvalidRange
	}
	parse := func(value string, fallback time.Time) (time.Time, error) {
		if value == "" {
			return fallback, nil
		}
		if parsed, parseErr := time.Parse("2006-01-02", value); parseErr == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, location).UTC(), nil
		}
		parsed, parseErr := time.Parse(time.RFC3339, value)
		if parseErr != nil {
			return time.Time{}, ErrInvalidRange
		}
		return parsed.UTC(), nil
	}
	defaultTo := now.UTC()
	defaultFrom := defaultTo.Add(-30 * 24 * time.Hour)
	start, err := parse(from, defaultFrom)
	if err != nil {
		return Range{}, err
	}
	end, err := parse(to, defaultTo)
	if err != nil {
		return Range{}, err
	}
	if !end.After(start) || end.Sub(start) > 366*24*time.Hour {
		return Range{}, ErrInvalidRange
	}
	return Range{From: start, To: end, Timezone: timezone}, nil
}
func where(filters Filters) (string, []any) {
	query := `organization_id=$1 AND request_timestamp >= $2 AND request_timestamp < $3 AND (NULLIF($4,'') IS NULL OR project_id=NULLIF($4,'')::uuid) AND (NULLIF($5,'') IS NULL OR workload_id=NULLIF($5,'')::uuid) AND ($6='' OR provider=$6) AND ($7='' OR model=$7) AND currency=$8`
	return query, []any{filters.Range.From, filters.Range.To, filters.ProjectID, filters.WorkloadID, filters.Provider, filters.Model, filters.Currency}
}
func (s *Store) Overview(ctx context.Context, orgID string, filters Filters) (Overview, error) {
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	var out Overview
	var cost, avg, errorRate, retryRate string
	query := fmt.Sprintf(`SELECT COALESCE(sum(estimated_cost),0)::text,count(*),COALESCE(sum(input_tokens),0),COALESCE(sum(output_tokens),0),COALESCE(sum(cached_input_tokens),0),COALESCE(avg(latency_ms),0)::text,COALESCE(count(*) FILTER(WHERE status<>'success')::numeric/NULLIF(count(*),0),0)::text,COALESCE(count(*) FILTER(WHERE retry_count>0)::numeric/NULLIF(count(*),0),0)::text,COALESCE(min(currency),'USD') FROM llm_calls WHERE %s`, clause)
	err := s.pool.QueryRow(ctx, query, args...).Scan(&cost, &out.CallCount, &out.InputTokens, &out.OutputTokens, &out.CachedInputTokens, &avg, &errorRate, &retryRate, &out.Currency)
	if err != nil {
		return Overview{}, err
	}
	out.ObservedCost = cost
	out.AverageLatencyMS = avg
	out.ErrorRate = errorRate
	out.RetryRate = retryRate
	duration := filters.Range.To.Sub(filters.Range.From)
	previous := filters
	previous.Range.From = filters.Range.From.Add(-duration)
	previous.Range.To = filters.Range.From
	previousClause, previousArgs := where(previous)
	previousArgs = append([]any{orgID}, previousArgs...)
	if err = s.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE(sum(estimated_cost),0)::text FROM llm_calls WHERE %s`, previousClause), previousArgs...).Scan(&out.PreviousPeriodCost); err != nil {
		return Overview{}, err
	}
	amount, err := decimal.NewFromString(cost)
	if err != nil {
		return Overview{}, err
	}
	days := decimal.NewFromFloat(duration.Hours() / 24)
	if days.IsZero() {
		return Overview{}, ErrInvalidRange
	}
	out.EstimatedMonthlyProjection = amount.Div(days).Mul(decimal.RequireFromString("30.4375")).StringFixed(12)
	out.ProjectionMethod = "observed range daily average × 30.4375 days"
	return out, nil
}
func (s *Store) Costs(ctx context.Context, orgID, dimension string, filters Filters) ([]CostPoint, error) {
	expression, ok := map[string]string{"project": "project_id::text", "workload": "workload_id::text", "provider": "provider", "model": "provider||'/'||model", "task": "COALESCE(NULLIF(metadata->>'task_type',''),'unknown')"}[dimension]
	if !ok {
		return nil, errors.New("invalid dimension")
	}
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	rows, err := s.pool.Query(ctx, fmt.Sprintf(`SELECT %s,COALESCE(sum(estimated_cost),0)::text,count(*),COALESCE(min(currency),'USD') FROM llm_calls WHERE %s GROUP BY 1 ORDER BY sum(estimated_cost) DESC LIMIT 100`, expression, clause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []CostPoint{}
	for rows.Next() {
		var item CostPoint
		if err = rows.Scan(&item.Key, &item.Cost, &item.Calls, &item.Currency); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Daily(ctx context.Context, orgID string, filters Filters) ([]DailyPoint, error) {
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	args = append(args, filters.Range.Timezone)
	rows, err := s.pool.Query(ctx, fmt.Sprintf(`SELECT to_char(date_trunc('day',request_timestamp AT TIME ZONE $9),'YYYY-MM-DD'),COALESCE(sum(estimated_cost),0)::text,count(*) FROM llm_calls WHERE %s GROUP BY 1 ORDER BY 1 LIMIT 367`, clause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []DailyPoint{}
	for rows.Next() {
		var item DailyPoint
		if err = rows.Scan(&item.Date, &item.Cost, &item.Calls); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Tokens(ctx context.Context, orgID string, filters Filters) (TokenSummary, error) {
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	var out TokenSummary
	err := s.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE(sum(input_tokens),0),COALESCE(sum(output_tokens),0),COALESCE(sum(cached_input_tokens),0) FROM llm_calls WHERE %s`, clause), args...).Scan(&out.Input, &out.Output, &out.CachedInput)
	return out, err
}
func (s *Store) Latency(ctx context.Context, orgID string, filters Filters) (LatencySummary, error) {
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	var out LatencySummary
	err := s.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE(avg(latency_ms),0)::text,COALESCE(percentile_cont(.5) WITHIN GROUP(ORDER BY latency_ms),0)::text,COALESCE(percentile_cont(.95) WITHIN GROUP(ORDER BY latency_ms),0)::text FROM llm_calls WHERE %s`, clause), args...).Scan(&out.AverageMS, &out.P50MS, &out.P95MS)
	return out, err
}
func (s *Store) Errors(ctx context.Context, orgID string, filters Filters) (ErrorSummary, error) {
	clause, args := where(filters)
	args = append([]any{orgID}, args...)
	var out ErrorSummary
	err := s.pool.QueryRow(ctx, fmt.Sprintf(`SELECT count(*),count(*) FILTER(WHERE status<>'success'),count(*) FILTER(WHERE retry_count>0),COALESCE(count(*) FILTER(WHERE status<>'success')::numeric/NULLIF(count(*),0),0)::text,COALESCE(count(*) FILTER(WHERE retry_count>0)::numeric/NULLIF(count(*),0),0)::text FROM llm_calls WHERE %s`, clause), args...).Scan(&out.Calls, &out.Errors, &out.RetriedCalls, &out.ErrorRate, &out.RetryRate)
	return out, err
}
