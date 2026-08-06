package ingestion

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/pricing"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/redaction"
)

var ErrInvalid = errors.New("invalid call")
var ErrConflict = errors.New("idempotency conflict")
var ErrNotFound = errors.New("resource or pricing not found")

type Input struct {
	ProjectID         string         `json:"project_id"`
	WorkloadID        string         `json:"workload_id"`
	ExternalCallID    string         `json:"external_call_id"`
	TraceID           string         `json:"trace_id"`
	Provider          string         `json:"provider"`
	Model             string         `json:"model"`
	PromptTemplateID  string         `json:"prompt_template_id"`
	RequestTimestamp  time.Time      `json:"request_timestamp"`
	ResponseTimestamp *time.Time     `json:"response_timestamp"`
	LatencyMS         int64          `json:"latency_ms"`
	InputTokens       int64          `json:"input_tokens"`
	OutputTokens      int64          `json:"output_tokens"`
	CachedInputTokens int64          `json:"cached_input_tokens"`
	RetryCount        int            `json:"retry_count"`
	Status            string         `json:"status"`
	ErrorCode         string         `json:"error_code"`
	Prompt            string         `json:"prompt"`
	Response          string         `json:"response"`
	Metadata          map[string]any `json:"metadata"`
}
type Call struct {
	ID                      string    `json:"id"`
	OrganizationID          string    `json:"organization_id"`
	ProjectID               string    `json:"project_id"`
	WorkloadID              string    `json:"workload_id"`
	ExternalCallID          string    `json:"external_call_id"`
	Provider                string    `json:"provider"`
	Model                   string    `json:"model"`
	EstimatedCost           string    `json:"estimated_cost"`
	Currency                string    `json:"currency"`
	RedactedPromptPreview   string    `json:"redacted_prompt_preview,omitempty"`
	RedactedResponsePreview string    `json:"redacted_response_preview,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	IdempotentReplay        bool      `json:"idempotent_replay"`
}
type CallPage struct {
	Items  []Call `json:"items"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
type prepared struct {
	Input
	Currency, Cost, PromptPreview, ResponsePreview string
	PromptHash                                     []byte
	PayloadHash                                    [32]byte
	Metadata                                       []byte
}
type Store struct {
	pool   *pgxpool.Pool
	prices *pricing.Store
}

func NewStore(pool *pgxpool.Pool, prices *pricing.Store) *Store {
	return &Store{pool: pool, prices: prices}
}
func Validate(input Input) error {
	if input.ProjectID == "" || input.WorkloadID == "" || strings.TrimSpace(input.ExternalCallID) == "" || input.Provider == "" || input.Model == "" || input.RequestTimestamp.IsZero() || input.LatencyMS < 0 || input.InputTokens < 0 || input.OutputTokens < 0 || input.CachedInputTokens < 0 || input.RetryCount < 0 || !oneOf(input.Status, "success", "error", "timeout") {
		return ErrInvalid
	}
	if input.ResponseTimestamp != nil && input.ResponseTimestamp.Before(input.RequestTimestamp) {
		return ErrInvalid
	}
	return nil
}
func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func (s *Store) Ingest(ctx context.Context, orgID string, projectRestriction *string, input Input) (Call, error) {
	calls, err := s.IngestBatch(ctx, orgID, projectRestriction, []Input{input})
	if err != nil {
		return Call{}, err
	}
	return calls[0], nil
}
func (s *Store) Get(ctx context.Context, orgID, id string) (Call, error) {
	var call Call
	err := s.pool.QueryRow(ctx, `SELECT id,organization_id,project_id,workload_id,external_call_id,provider,model,estimated_cost::text,currency,COALESCE(redacted_prompt_preview,''),COALESCE(redacted_response_preview,''),created_at FROM llm_calls WHERE organization_id=$1 AND id=$2`, orgID, id).Scan(&call.ID, &call.OrganizationID, &call.ProjectID, &call.WorkloadID, &call.ExternalCallID, &call.Provider, &call.Model, &call.EstimatedCost, &call.Currency, &call.RedactedPromptPreview, &call.RedactedResponsePreview, &call.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Call{}, ErrNotFound
	}
	return call, err
}
func (s *Store) List(ctx context.Context, orgID, projectID, workloadID string, limit, offset int) (CallPage, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,project_id,workload_id,external_call_id,provider,model,estimated_cost::text,currency,COALESCE(redacted_prompt_preview,''),COALESCE(redacted_response_preview,''),created_at FROM llm_calls WHERE organization_id=$1 AND (NULLIF($2,'') IS NULL OR project_id=NULLIF($2,'')::uuid) AND (NULLIF($3,'') IS NULL OR workload_id=NULLIF($3,'')::uuid) ORDER BY request_timestamp DESC,id LIMIT $4 OFFSET $5`, orgID, projectID, workloadID, limit, offset)
	if err != nil {
		return CallPage{}, err
	}
	defer rows.Close()
	page := CallPage{Items: []Call{}, Limit: limit, Offset: offset}
	for rows.Next() {
		var call Call
		if err = rows.Scan(&call.ID, &call.OrganizationID, &call.ProjectID, &call.WorkloadID, &call.ExternalCallID, &call.Provider, &call.Model, &call.EstimatedCost, &call.Currency, &call.RedactedPromptPreview, &call.RedactedResponsePreview, &call.CreatedAt); err != nil {
			return CallPage{}, err
		}
		page.Items = append(page.Items, call)
	}
	return page, rows.Err()
}
func (s *Store) IngestBatch(ctx context.Context, orgID string, projectRestriction *string, inputs []Input) ([]Call, error) {
	if len(inputs) == 0 || len(inputs) > 100 {
		return nil, ErrInvalid
	}
	preparedCalls := make([]prepared, len(inputs))
	seen := map[string]bool{}
	for i, input := range inputs {
		if Validate(input) != nil || seen[input.ExternalCallID] || (projectRestriction != nil && *projectRestriction != input.ProjectID) {
			return nil, ErrInvalid
		}
		seen[input.ExternalCallID] = true
		item, err := s.prepare(ctx, orgID, input)
		if err != nil {
			return nil, err
		}
		preparedCalls[i] = item
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	result := make([]Call, len(inputs))
	for i, item := range preparedCalls {
		call, insertErr := insert(ctx, tx, orgID, item)
		if insertErr != nil {
			return nil, insertErr
		}
		result[i] = call
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
func (s *Store) prepare(ctx context.Context, orgID string, input Input) (prepared, error) {
	var currency string
	err := s.pool.QueryRow(ctx, `SELECT p.currency FROM projects p JOIN workloads w ON w.project_id=p.id WHERE p.organization_id=$1 AND p.id=$2 AND w.id=$3`, orgID, input.ProjectID, input.WorkloadID).Scan(&currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return prepared{}, ErrNotFound
	}
	if err != nil {
		return prepared{}, err
	}
	price, err := s.prices.EffectivePrice(ctx, input.Provider, input.Model, currency, input.RequestTimestamp)
	if errors.Is(err, pricing.ErrNotFound) {
		return prepared{}, ErrNotFound
	}
	if err != nil {
		return prepared{}, err
	}
	cost, err := pricing.Calculate(price, pricing.Usage{InputTokens: input.InputTokens, OutputTokens: input.OutputTokens, CachedInputTokens: input.CachedInputTokens})
	if err != nil {
		return prepared{}, ErrInvalid
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return prepared{}, ErrInvalid
	}
	cleanMetadata, err := json.Marshal(redaction.Metadata(input.Metadata, true))
	if err != nil {
		return prepared{}, ErrInvalid
	}
	item := prepared{Input: input, Currency: currency, Cost: cost.Total.StringFixed(12), PromptPreview: redaction.Preview(input.Prompt, 512), ResponsePreview: redaction.Preview(input.Response, 512), PayloadHash: sha256.Sum256(payload), Metadata: cleanMetadata}
	if input.Prompt != "" {
		hash := sha256.Sum256([]byte(input.Prompt))
		item.PromptHash = hash[:]
	}
	item.Prompt = ""
	item.Response = ""
	return item, nil
}

type dbtx interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func insert(ctx context.Context, tx dbtx, orgID string, item prepared) (Call, error) {
	var call Call
	var storedHash []byte
	err := tx.QueryRow(ctx, `INSERT INTO llm_calls(organization_id,project_id,workload_id,external_call_id,trace_id,provider,model,prompt_template_id,request_timestamp,response_timestamp,latency_ms,input_tokens,output_tokens,cached_input_tokens,retry_count,status,error_code,estimated_cost,currency,prompt_hash,redacted_prompt_preview,redacted_response_preview,metadata,payload_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24) ON CONFLICT(organization_id,external_call_id) DO UPDATE SET external_call_id=EXCLUDED.external_call_id RETURNING id,organization_id,project_id,workload_id,external_call_id,provider,model,estimated_cost::text,currency,redacted_prompt_preview,redacted_response_preview,created_at,payload_hash,(xmax<>0)`, orgID, item.ProjectID, item.WorkloadID, item.ExternalCallID, nullText(item.TraceID), item.Provider, item.Model, nullText(item.PromptTemplateID), item.RequestTimestamp, item.ResponseTimestamp, item.LatencyMS, item.InputTokens, item.OutputTokens, item.CachedInputTokens, item.RetryCount, item.Status, nullText(item.ErrorCode), item.Cost, item.Currency, item.PromptHash, nullText(item.PromptPreview), nullText(item.ResponsePreview), item.Metadata, item.PayloadHash[:]).Scan(&call.ID, &call.OrganizationID, &call.ProjectID, &call.WorkloadID, &call.ExternalCallID, &call.Provider, &call.Model, &call.EstimatedCost, &call.Currency, &call.RedactedPromptPreview, &call.RedactedResponsePreview, &call.CreatedAt, &storedHash, &call.IdempotentReplay)
	if err != nil {
		return Call{}, err
	}
	if subtle.ConstantTimeCompare(storedHash, item.PayloadHash[:]) != 1 {
		return Call{}, ErrConflict
	}
	return call, nil
}
func nullText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
