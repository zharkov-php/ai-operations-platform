package portfolio

import (
	"context"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Environment    string    `json:"environment"`
	MonthlyBudget  string    `json:"monthly_budget"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type Workload struct {
	ID                    string    `json:"id"`
	ProjectID             string    `json:"project_id"`
	Name                  string    `json:"name"`
	Slug                  string    `json:"slug"`
	Type                  string    `json:"type"`
	Owner                 string    `json:"owner"`
	Criticality           string    `json:"criticality"`
	QualityRequirement    string    `json:"quality_requirement"`
	PrivacyClassification string    `json:"privacy_classification"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
type ProjectInput struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Environment   string `json:"environment"`
	MonthlyBudget string `json:"monthly_budget"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}
type WorkloadInput struct {
	ProjectID             string `json:"project_id"`
	Name                  string `json:"name"`
	Slug                  string `json:"slug"`
	Type                  string `json:"type"`
	Owner                 string `json:"owner"`
	Criticality           string `json:"criticality"`
	QualityRequirement    string `json:"quality_requirement"`
	PrivacyClassification string `json:"privacy_classification"`
	Status                string `json:"status"`
}
type ListOptions struct {
	Limit, Offset                  int
	Status, Environment, ProjectID string
}
type Page[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func ValidateProject(v ProjectInput) error {
	if strings.TrimSpace(v.Name) == "" || !slugPattern.MatchString(v.Slug) {
		return errors.New("name and valid slug are required")
	}
	if !oneOf(v.Environment, "development", "staging", "production") || !regexp.MustCompile(`^[A-Za-z]{3}$`).MatchString(v.Currency) {
		return errors.New("invalid environment or currency")
	}
	amount, ok := new(big.Rat).SetString(v.MonthlyBudget)
	if !ok || amount.Sign() < 0 {
		return errors.New("monthly_budget must be a non-negative decimal")
	}
	if v.Status != "" && !oneOf(v.Status, "active", "archived") {
		return errors.New("invalid status")
	}
	return nil
}
func ValidateWorkload(v WorkloadInput) error {
	if v.ProjectID == "" || strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Owner) == "" || !slugPattern.MatchString(v.Slug) {
		return errors.New("project_id, name, owner, and valid slug are required")
	}
	if !oneOf(v.Type, "agent", "feature", "workflow", "scheduled_job", "internal_tool") || !oneOf(v.Criticality, "low", "medium", "high", "critical") || !oneOf(v.QualityRequirement, "standard", "high", "critical") || !oneOf(v.PrivacyClassification, "public", "internal", "confidential", "restricted") {
		return errors.New("invalid workload classification")
	}
	if v.Status != "" && !oneOf(v.Status, "active", "archived") {
		return errors.New("invalid status")
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
func normalize(opts ListOptions) ListOptions {
	if opts.Limit <= 0 {
		opts.Limit = 25
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	return opts
}

func (s *Store) CreateProject(ctx context.Context, orgID, actorID string, in ProjectInput) (Project, error) {
	if in.Status == "" {
		in.Status = "active"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Project{}, err
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency,status) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, orgID, in.Name, in.Slug, in.Environment, in.MonthlyBudget, strings.ToUpper(in.Currency), in.Status).Scan(&id)
	if err != nil {
		return Project{}, mapError(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id) VALUES($1,$2,'project.created','project',$3)`, orgID, actorID, id); err != nil {
		return Project{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Project{}, err
	}
	return s.GetProject(ctx, orgID, id)
}
func (s *Store) GetProject(ctx context.Context, orgID, id string) (Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx, `SELECT id,organization_id,name,slug,environment,monthly_budget::text,currency,status,created_at,updated_at FROM projects WHERE organization_id=$1 AND id=$2`, orgID, id).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Environment, &p.MonthlyBudget, &p.Currency, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return p, err
}
func (s *Store) ListProjects(ctx context.Context, orgID string, opts ListOptions) (Page[Project], error) {
	opts = normalize(opts)
	rows, err := s.pool.Query(ctx, `SELECT id,organization_id,name,slug,environment,monthly_budget::text,currency,status,created_at,updated_at FROM projects WHERE organization_id=$1 AND ($2='' OR status=$2) AND ($3='' OR environment=$3) ORDER BY id LIMIT $4 OFFSET $5`, orgID, opts.Status, opts.Environment, opts.Limit, opts.Offset)
	if err != nil {
		return Page[Project]{}, err
	}
	defer rows.Close()
	page := Page[Project]{Items: []Project{}, Limit: opts.Limit, Offset: opts.Offset}
	for rows.Next() {
		var p Project
		if err = rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Environment, &p.MonthlyBudget, &p.Currency, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return Page[Project]{}, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}
func (s *Store) UpdateProject(ctx context.Context, orgID, actorID, id string, in ProjectInput) (Project, error) {
	command, err := s.pool.Exec(ctx, `UPDATE projects SET name=$3,slug=$4,environment=$5,monthly_budget=$6,currency=$7,status=$8,updated_at=CURRENT_TIMESTAMP WHERE organization_id=$1 AND id=$2`, orgID, id, in.Name, in.Slug, in.Environment, in.MonthlyBudget, strings.ToUpper(in.Currency), in.Status)
	if err != nil {
		return Project{}, mapError(err)
	}
	if command.RowsAffected() == 0 {
		return Project{}, ErrNotFound
	}
	_, _ = s.pool.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id) VALUES($1,$2,'project.updated','project',$3)`, orgID, actorID, id)
	return s.GetProject(ctx, orgID, id)
}

func (s *Store) CreateWorkload(ctx context.Context, orgID, actorID string, in WorkloadInput) (Workload, error) {
	if in.Status == "" {
		in.Status = "active"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Workload{}, err
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification,status) SELECT id,$3,$4,$5,$6,$7,$8,$9,$10 FROM projects WHERE id=$2 AND organization_id=$1 RETURNING id`, orgID, in.ProjectID, in.Name, in.Slug, in.Type, in.Owner, in.Criticality, in.QualityRequirement, in.PrivacyClassification, in.Status).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workload{}, ErrNotFound
	}
	if err != nil {
		return Workload{}, mapError(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id) VALUES($1,$2,'workload.created','workload',$3)`, orgID, actorID, id); err != nil {
		return Workload{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Workload{}, err
	}
	return s.GetWorkload(ctx, orgID, id)
}
func (s *Store) GetWorkload(ctx context.Context, orgID, id string) (Workload, error) {
	var w Workload
	err := s.pool.QueryRow(ctx, `SELECT w.id,w.project_id,w.name,w.slug,w.type,w.owner,w.criticality,w.quality_requirement,w.privacy_classification,w.status,w.created_at,w.updated_at FROM workloads w JOIN projects p ON p.id=w.project_id WHERE p.organization_id=$1 AND w.id=$2`, orgID, id).Scan(&w.ID, &w.ProjectID, &w.Name, &w.Slug, &w.Type, &w.Owner, &w.Criticality, &w.QualityRequirement, &w.PrivacyClassification, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workload{}, ErrNotFound
	}
	return w, err
}
func (s *Store) ListWorkloads(ctx context.Context, orgID string, opts ListOptions) (Page[Workload], error) {
	opts = normalize(opts)
	rows, err := s.pool.Query(ctx, `SELECT w.id,w.project_id,w.name,w.slug,w.type,w.owner,w.criticality,w.quality_requirement,w.privacy_classification,w.status,w.created_at,w.updated_at FROM workloads w JOIN projects p ON p.id=w.project_id WHERE p.organization_id=$1 AND ($2='' OR w.status=$2) AND (NULLIF($3,'') IS NULL OR w.project_id=NULLIF($3,'')::uuid) ORDER BY w.id LIMIT $4 OFFSET $5`, orgID, opts.Status, opts.ProjectID, opts.Limit, opts.Offset)
	if err != nil {
		return Page[Workload]{}, err
	}
	defer rows.Close()
	page := Page[Workload]{Items: []Workload{}, Limit: opts.Limit, Offset: opts.Offset}
	for rows.Next() {
		var w Workload
		if err = rows.Scan(&w.ID, &w.ProjectID, &w.Name, &w.Slug, &w.Type, &w.Owner, &w.Criticality, &w.QualityRequirement, &w.PrivacyClassification, &w.Status, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return Page[Workload]{}, err
		}
		page.Items = append(page.Items, w)
	}
	return page, rows.Err()
}
func (s *Store) UpdateWorkload(ctx context.Context, orgID, actorID, id string, in WorkloadInput) (Workload, error) {
	command, err := s.pool.Exec(ctx, `UPDATE workloads w SET project_id=$3,name=$4,slug=$5,type=$6,owner=$7,criticality=$8,quality_requirement=$9,privacy_classification=$10,status=$11,updated_at=CURRENT_TIMESTAMP FROM projects current_project,projects target_project WHERE w.id=$2 AND current_project.id=w.project_id AND current_project.organization_id=$1 AND target_project.id=$3 AND target_project.organization_id=$1`, orgID, id, in.ProjectID, in.Name, in.Slug, in.Type, in.Owner, in.Criticality, in.QualityRequirement, in.PrivacyClassification, in.Status)
	if err != nil {
		return Workload{}, mapError(err)
	}
	if command.RowsAffected() == 0 {
		return Workload{}, ErrNotFound
	}
	_, _ = s.pool.Exec(ctx, `INSERT INTO audit_entries(organization_id,actor_user_id,action,subject_type,subject_id) VALUES($1,$2,'workload.updated','workload',$3)`, orgID, actorID, id)
	return s.GetWorkload(ctx, orgID, id)
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
