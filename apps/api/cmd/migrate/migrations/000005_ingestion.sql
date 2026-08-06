CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID REFERENCES projects(id), name TEXT NOT NULL, key_prefix TEXT NOT NULL, key_hash BYTEA NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL, last_used_at TIMESTAMPTZ, revoked_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX api_keys_org_idx ON api_keys(organization_id,created_at DESC);

CREATE TABLE llm_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organization_id UUID NOT NULL REFERENCES organizations(id),
    project_id UUID NOT NULL REFERENCES projects(id), workload_id UUID NOT NULL REFERENCES workloads(id),
    external_call_id TEXT NOT NULL, trace_id TEXT, provider TEXT NOT NULL, model TEXT NOT NULL,
    prompt_template_id TEXT, request_timestamp TIMESTAMPTZ NOT NULL, response_timestamp TIMESTAMPTZ,
    latency_ms BIGINT NOT NULL CHECK(latency_ms>=0), input_tokens BIGINT NOT NULL CHECK(input_tokens>=0),
    output_tokens BIGINT NOT NULL CHECK(output_tokens>=0), cached_input_tokens BIGINT NOT NULL DEFAULT 0 CHECK(cached_input_tokens>=0),
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK(retry_count>=0), status TEXT NOT NULL CHECK(status IN('success','error','timeout')),
    error_code TEXT, estimated_cost NUMERIC(24,12) NOT NULL CHECK(estimated_cost>=0), currency CHAR(3) NOT NULL,
    prompt_hash BYTEA, redacted_prompt_preview TEXT, redacted_response_preview TEXT,
    metadata JSONB NOT NULL DEFAULT '{}', payload_hash BYTEA NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(provider,model) REFERENCES provider_models(provider,model),
    UNIQUE(organization_id,external_call_id)
);
CREATE INDEX llm_calls_org_time_idx ON llm_calls(organization_id,request_timestamp DESC,id);
CREATE INDEX llm_calls_project_time_idx ON llm_calls(project_id,request_timestamp DESC);
CREATE INDEX llm_calls_workload_time_idx ON llm_calls(workload_id,request_timestamp DESC);
CREATE INDEX llm_calls_prompt_hash_idx ON llm_calls(organization_id,prompt_hash) WHERE prompt_hash IS NOT NULL;
