package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("invalid API key")
var allowedScopes = map[string]bool{"ingest:calls": true, "read:analytics": true, "read:recommendations": true, "manage:recommendations": true, "manage:evaluations": true, "manage:experiments": true, "manage:pricing": true, "manage:team": true}

type Key struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	ProjectID      *string    `json:"project_id"`
	Name           string     `json:"name"`
	Prefix         string     `json:"prefix"`
	Scopes         []string   `json:"scopes"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	RevokedAt      *time.Time `json:"revoked_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
type Created struct {
	Key
	Secret string `json:"secret"`
}
type Principal struct {
	KeyID, OrganizationID string
	ProjectID             *string
	Scopes                map[string]bool
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func Validate(name string, scopes []string) error {
	if strings.TrimSpace(name) == "" || len(scopes) == 0 {
		return errors.New("name and scopes are required")
	}
	seen := map[string]bool{}
	for _, scope := range scopes {
		if !allowedScopes[scope] || seen[scope] {
			return errors.New("invalid or duplicate scope")
		}
		seen[scope] = true
	}
	return nil
}
func (s *Store) Create(ctx context.Context, orgID, userID, name string, projectID *string, scopes []string) (Created, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Created{}, err
	}
	random := base64.RawURLEncoding.EncodeToString(raw)
	secret := "aop_" + random
	prefix := secret[:12]
	hash := sha256.Sum256([]byte(secret))
	var key Key
	err := s.pool.QueryRow(ctx, `INSERT INTO api_keys(organization_id,project_id,name,key_prefix,key_hash,scopes,created_by) SELECT $1,p.id,$3,$4,$5,$6,$7 FROM (SELECT $2::uuid AS requested) x LEFT JOIN projects p ON p.id=x.requested AND p.organization_id=$1 WHERE $2::uuid IS NULL OR p.id IS NOT NULL RETURNING id,organization_id,project_id,name,key_prefix,scopes,last_used_at,revoked_at,created_at`, orgID, projectID, name, prefix, hash[:], scopes, userID).Scan(&key.ID, &key.OrganizationID, &key.ProjectID, &key.Name, &key.Prefix, &key.Scopes, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Created{}, errors.New("project not found")
	}
	return Created{Key: key, Secret: secret}, err
}
func (s *Store) List(ctx context.Context, orgID string) ([]Key, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,project_id,name,key_prefix,scopes,last_used_at,revoked_at,created_at FROM api_keys WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 100`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Key{}
	for rows.Next() {
		var key Key
		if err = rows.Scan(&key.ID, &key.OrganizationID, &key.ProjectID, &key.Name, &key.Prefix, &key.Scopes, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, key)
	}
	return items, rows.Err()
}
func (s *Store) Revoke(ctx context.Context, orgID, id string) error {
	command, err := s.pool.Exec(ctx, `UPDATE api_keys SET revoked_at=COALESCE(revoked_at,CURRENT_TIMESTAMP) WHERE organization_id=$1 AND id=$2`, orgID, id)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrInvalid
	}
	return nil
}
func (s *Store) Authenticate(ctx context.Context, secret string) (Principal, error) {
	if !strings.HasPrefix(secret, "aop_") {
		return Principal{}, ErrInvalid
	}
	hash := sha256.Sum256([]byte(secret))
	var principal Principal
	var scopes []string
	err := s.pool.QueryRow(ctx, `UPDATE api_keys SET last_used_at=CURRENT_TIMESTAMP WHERE key_hash=$1 AND revoked_at IS NULL RETURNING id,organization_id,project_id,scopes`, hash[:]).Scan(&principal.KeyID, &principal.OrganizationID, &principal.ProjectID, &scopes)
	if err != nil {
		return Principal{}, ErrInvalid
	}
	principal.Scopes = map[string]bool{}
	for _, scope := range scopes {
		principal.Scopes[scope] = true
	}
	return principal, nil
}
