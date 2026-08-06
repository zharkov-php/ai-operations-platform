CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organization_id UUID NOT NULL REFERENCES organizations(id),
    name TEXT NOT NULL, slug TEXT NOT NULL, environment TEXT NOT NULL CHECK (environment IN ('development','staging','production')),
    monthly_budget NUMERIC(20,6) NOT NULL DEFAULT 0 CHECK (monthly_budget >= 0), currency CHAR(3) NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, slug)
);
CREATE TABLE workloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL, slug TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('agent','feature','workflow','scheduled_job','internal_tool')),
    owner TEXT NOT NULL, criticality TEXT NOT NULL CHECK (criticality IN ('low','medium','high','critical')),
    quality_requirement TEXT NOT NULL CHECK (quality_requirement IN ('standard','high','critical')),
    privacy_classification TEXT NOT NULL CHECK (privacy_classification IN ('public','internal','confidential','restricted')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, slug)
);
CREATE INDEX projects_org_status_idx ON projects (organization_id,status,id);
CREATE INDEX workloads_project_status_idx ON workloads (project_id,status,id);
