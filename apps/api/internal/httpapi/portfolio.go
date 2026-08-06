package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/auth"
	"github.com/zharkov-php/ai-operations-platform/apps/api/internal/portfolio"
)

func registerPortfolioRoutes(router chi.Router, authService *auth.Service, store *portfolio.Store) {
	if authService == nil || store == nil {
		return
	}
	router.Group(func(private chi.Router) {
		private.Use(authenticate(authService))
		private.Get("/api/v1/projects", listProjects(store))
		private.Get("/api/v1/projects/{projectID}", getProject(store))
		private.Get("/api/v1/workloads", listWorkloads(store))
		private.Get("/api/v1/workloads/{workloadID}", getWorkload(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/projects", createProject(store))
		private.With(requireRoles("owner", "admin", "engineer")).Patch("/api/v1/projects/{projectID}", updateProject(store))
		private.With(requireRoles("owner", "admin", "engineer")).Post("/api/v1/workloads", createWorkload(store))
		private.With(requireRoles("owner", "admin", "engineer")).Patch("/api/v1/workloads/{workloadID}", updateWorkload(store))
	})
}
func requestClaims(r *http.Request) auth.Claims { return r.Context().Value(claimsKey{}).(auth.Claims) }
func listOptions(r *http.Request) portfolio.ListOptions {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return portfolio.ListOptions{Limit: limit, Offset: offset, Status: r.URL.Query().Get("status"), Environment: r.URL.Query().Get("environment"), ProjectID: r.URL.Query().Get("project_id")}
}
func createProject(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in portfolio.ProjectInput
		if json.NewDecoder(r.Body).Decode(&in) != nil || portfolio.ValidateProject(in) != nil {
			writeError(w, r, 400, "invalid_project", "project payload is invalid")
			return
		}
		c := requestClaims(r)
		result, err := store.CreateProject(r.Context(), c.OrganizationID, c.UserID, in)
		writePortfolioResult(w, r, result, err, http.StatusCreated)
	}
}
func listProjects(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.ListProjects(r.Context(), requestClaims(r).OrganizationID, listOptions(r))
		writePortfolioResult(w, r, result, err, 200)
	}
}
func getProject(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.GetProject(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "projectID"))
		writePortfolioResult(w, r, result, err, 200)
	}
}

type projectPatch struct {
	Name          *string `json:"name"`
	Slug          *string `json:"slug"`
	Environment   *string `json:"environment"`
	MonthlyBudget *string `json:"monthly_budget"`
	Currency      *string `json:"currency"`
	Status        *string `json:"status"`
}

func updateProject(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := requestClaims(r)
		current, err := store.GetProject(r.Context(), c.OrganizationID, chi.URLParam(r, "projectID"))
		if err != nil {
			writePortfolioResult(w, r, current, err, 200)
			return
		}
		var patch projectPatch
		if json.NewDecoder(r.Body).Decode(&patch) != nil {
			writeError(w, r, 400, "invalid_project", "project payload is invalid")
			return
		}
		in := portfolio.ProjectInput{Name: current.Name, Slug: current.Slug, Environment: current.Environment, MonthlyBudget: current.MonthlyBudget, Currency: current.Currency, Status: current.Status}
		if patch.Name != nil {
			in.Name = *patch.Name
		}
		if patch.Slug != nil {
			in.Slug = *patch.Slug
		}
		if patch.Environment != nil {
			in.Environment = *patch.Environment
		}
		if patch.MonthlyBudget != nil {
			in.MonthlyBudget = *patch.MonthlyBudget
		}
		if patch.Currency != nil {
			in.Currency = *patch.Currency
		}
		if patch.Status != nil {
			in.Status = *patch.Status
		}
		if portfolio.ValidateProject(in) != nil {
			writeError(w, r, 400, "invalid_project", "project payload is invalid")
			return
		}
		result, err := store.UpdateProject(r.Context(), c.OrganizationID, c.UserID, current.ID, in)
		writePortfolioResult(w, r, result, err, 200)
	}
}
func createWorkload(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in portfolio.WorkloadInput
		if json.NewDecoder(r.Body).Decode(&in) != nil || portfolio.ValidateWorkload(in) != nil {
			writeError(w, r, 400, "invalid_workload", "workload payload is invalid")
			return
		}
		c := requestClaims(r)
		result, err := store.CreateWorkload(r.Context(), c.OrganizationID, c.UserID, in)
		writePortfolioResult(w, r, result, err, 201)
	}
}
func listWorkloads(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.ListWorkloads(r.Context(), requestClaims(r).OrganizationID, listOptions(r))
		writePortfolioResult(w, r, result, err, 200)
	}
}
func getWorkload(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.GetWorkload(r.Context(), requestClaims(r).OrganizationID, chi.URLParam(r, "workloadID"))
		writePortfolioResult(w, r, result, err, 200)
	}
}

type workloadPatch struct {
	Name                  *string `json:"name"`
	Slug                  *string `json:"slug"`
	ProjectID             *string `json:"project_id"`
	Type                  *string `json:"type"`
	Owner                 *string `json:"owner"`
	Criticality           *string `json:"criticality"`
	QualityRequirement    *string `json:"quality_requirement"`
	PrivacyClassification *string `json:"privacy_classification"`
	Status                *string `json:"status"`
}

func updateWorkload(store *portfolio.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := requestClaims(r)
		current, err := store.GetWorkload(r.Context(), c.OrganizationID, chi.URLParam(r, "workloadID"))
		if err != nil {
			writePortfolioResult(w, r, current, err, 200)
			return
		}
		var patch workloadPatch
		if json.NewDecoder(r.Body).Decode(&patch) != nil {
			writeError(w, r, 400, "invalid_workload", "workload payload is invalid")
			return
		}
		in := portfolio.WorkloadInput{ProjectID: current.ProjectID, Name: current.Name, Slug: current.Slug, Type: current.Type, Owner: current.Owner, Criticality: current.Criticality, QualityRequirement: current.QualityRequirement, PrivacyClassification: current.PrivacyClassification, Status: current.Status}
		if patch.Name != nil {
			in.Name = *patch.Name
		}
		if patch.Slug != nil {
			in.Slug = *patch.Slug
		}
		if patch.ProjectID != nil {
			in.ProjectID = *patch.ProjectID
		}
		if patch.Type != nil {
			in.Type = *patch.Type
		}
		if patch.Owner != nil {
			in.Owner = *patch.Owner
		}
		if patch.Criticality != nil {
			in.Criticality = *patch.Criticality
		}
		if patch.QualityRequirement != nil {
			in.QualityRequirement = *patch.QualityRequirement
		}
		if patch.PrivacyClassification != nil {
			in.PrivacyClassification = *patch.PrivacyClassification
		}
		if patch.Status != nil {
			in.Status = *patch.Status
		}
		if portfolio.ValidateWorkload(in) != nil {
			writeError(w, r, 400, "invalid_workload", "workload payload is invalid")
			return
		}
		result, err := store.UpdateWorkload(r.Context(), c.OrganizationID, c.UserID, current.ID, in)
		writePortfolioResult(w, r, result, err, 200)
	}
}
func writePortfolioResult(w http.ResponseWriter, r *http.Request, result any, err error, status int) {
	switch {
	case err == nil:
		writeJSON(w, status, result)
	case errors.Is(err, portfolio.ErrNotFound):
		writeError(w, r, 404, "not_found", "resource not found")
	case errors.Is(err, portfolio.ErrConflict):
		writeError(w, r, 409, "conflict", "slug already exists")
	default:
		writeError(w, r, 500, "internal_error", "request could not be completed")
	}
}
