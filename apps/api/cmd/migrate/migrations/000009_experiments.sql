CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workload_id UUID NOT NULL REFERENCES workloads(id),
    recommendation_id UUID REFERENCES recommendations(id),
    control_execution JSONB NOT NULL,
    candidate_execution JSONB NOT NULL,
    traffic_percentage NUMERIC(5,2) NOT NULL CHECK(traffic_percentage > 0 AND traffic_percentage <= 100),
    status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN('draft','approved','running','paused','rolled_back','completed','verified')),
    guardrails JSONB NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    rollback_reason TEXT NOT NULL DEFAULT '',
    results JSONB NOT NULL DEFAULT '{}',
    verified_savings NUMERIC(24,12),
    currency CHAR(3) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX experiments_workload_idx ON experiments(workload_id, status, created_at DESC);
