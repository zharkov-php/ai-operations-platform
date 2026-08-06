CREATE TABLE evaluation_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workload_id UUID NOT NULL REFERENCES workloads(id),
    name TEXT NOT NULL CHECK(char_length(name) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL CHECK(char_length(source) BETWEEN 1 AND 100),
    privacy_classification TEXT NOT NULL CHECK(privacy_classification IN('public','internal','confidential','restricted')),
    case_count INTEGER NOT NULL DEFAULT 0 CHECK(case_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workload_id, name)
);

CREATE TABLE evaluation_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES evaluation_datasets(id) ON DELETE CASCADE,
    sanitized_input JSONB NOT NULL,
    expected_output JSONB NOT NULL,
    validation_rules JSONB NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE evaluation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES evaluation_datasets(id),
    candidate_execution JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN('pending','running','completed','failed')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    total_cases INTEGER NOT NULL DEFAULT 0,
    passed_cases INTEGER NOT NULL DEFAULT 0,
    failed_cases INTEGER NOT NULL DEFAULT 0,
    average_latency_ms NUMERIC(18,6) NOT NULL DEFAULT 0,
    estimated_cost NUMERIC(24,12) NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    results JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX evaluation_datasets_workload_idx ON evaluation_datasets(workload_id, created_at DESC);
CREATE INDEX evaluation_cases_dataset_idx ON evaluation_cases(dataset_id, created_at, id);
CREATE INDEX evaluation_runs_dataset_idx ON evaluation_runs(dataset_id, created_at DESC);
