CREATE TABLE local_model_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    hardware_name TEXT NOT NULL CHECK(char_length(hardware_name) BETWEEN 1 AND 200),
    purchase_cost NUMERIC(24,12) NOT NULL CHECK(purchase_cost >= 0),
    useful_lifetime_months INTEGER NOT NULL CHECK(useful_lifetime_months > 0),
    monthly_electricity NUMERIC(24,12) NOT NULL CHECK(monthly_electricity >= 0),
    monthly_maintenance NUMERIC(24,12) NOT NULL CHECK(monthly_maintenance >= 0),
    available_memory_gb NUMERIC(12,4) NOT NULL CHECK(available_memory_gb > 0),
    estimated_requests_per_second NUMERIC(18,6) NOT NULL CHECK(estimated_requests_per_second > 0),
    utilization NUMERIC(7,6) NOT NULL CHECK(utilization > 0 AND utilization <= 1),
    supported_model TEXT NOT NULL,
    context_limit INTEGER NOT NULL CHECK(context_limit > 0),
    currency CHAR(3) NOT NULL,
    benchmark_source TEXT NOT NULL DEFAULT 'user estimate',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX local_model_configurations_org_idx ON local_model_configurations(organization_id, created_at DESC);
