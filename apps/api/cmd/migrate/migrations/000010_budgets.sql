CREATE TABLE project_budget_thresholds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    percentage NUMERIC(5,2) NOT NULL CHECK(percentage > 0 AND percentage <= 200), severity TEXT NOT NULL CHECK(severity IN('info','warning','critical')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(project_id, percentage)
);
CREATE TABLE budget_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    threshold_type TEXT NOT NULL CHECK(threshold_type IN('percentage','projected_overspend','anomaly')), threshold_value NUMERIC(24,12) NOT NULL,
    severity TEXT NOT NULL CHECK(severity IN('info','warning','critical')), status TEXT NOT NULL DEFAULT 'open' CHECK(status IN('open','acknowledged','resolved')),
    dedupe_key TEXT NOT NULL, evidence JSONB NOT NULL DEFAULT '{}', triggered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMPTZ, acknowledged_by UUID REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX budget_alerts_open_dedupe_idx ON budget_alerts(project_id,dedupe_key) WHERE status='open';
CREATE INDEX budget_alerts_project_idx ON budget_alerts(project_id,status,triggered_at DESC);
