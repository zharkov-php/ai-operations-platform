CREATE TABLE task_classifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), workload_id UUID NOT NULL REFERENCES workloads(id),
    task_type TEXT NOT NULL CHECK(task_type IN('calculation','validation','classification','extraction','translation','summarization','generation','reasoning','coding','search','tool_selection','format_conversion','unknown')),
    determinism TEXT NOT NULL CHECK(determinism IN('low','medium','high','unknown')),
    complexity TEXT NOT NULL CHECK(complexity IN('low','medium','high','unknown')),
    privacy_classification TEXT NOT NULL, quality_requirement TEXT NOT NULL,
    failure_impact TEXT NOT NULL CHECK(failure_impact IN('low','medium','high','critical','unknown')),
    external_knowledge_required BOOLEAN NOT NULL, natural_language_understanding_required BOOLEAN NOT NULL,
    confidence NUMERIC(5,4) NOT NULL CHECK(confidence>=0 AND confidence<=1), classification_source TEXT NOT NULL,
    evidence JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workload_id,classification_source)
);
CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), workload_id UUID NOT NULL REFERENCES workloads(id),
    recommendation_type TEXT NOT NULL CHECK(recommendation_type IN('replace_with_code','exact_cache','smaller_hosted_model','local_model','reduce_context','reduce_output','structured_output','batch_requests','keep_current_model','manual_review_required')),
    priority TEXT NOT NULL CHECK(priority IN('low','medium','high')), confidence_level TEXT NOT NULL CHECK(confidence_level IN('low','medium','high')),
    confidence NUMERIC(5,4) NOT NULL CHECK(confidence>=0 AND confidence<=1), status TEXT NOT NULL DEFAULT 'new' CHECK(status IN('new','under_review','accepted','rejected','testing','applied','verified','dismissed')),
    current_execution JSONB NOT NULL, proposed_execution JSONB NOT NULL, reason_codes TEXT[] NOT NULL,
    evidence_summary JSONB NOT NULL, confidence_inputs JSONB NOT NULL,
    estimated_monthly_savings NUMERIC(24,12) NOT NULL DEFAULT 0, currency CHAR(3) NOT NULL,
    estimated_implementation_cost NUMERIC(24,12) NOT NULL DEFAULT 0, estimated_break_even_months NUMERIC(12,4),
    quality_risk TEXT NOT NULL, operational_risk TEXT NOT NULL, required_next_action TEXT NOT NULL,
    rule_version TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workload_id,recommendation_type,rule_version)
);
CREATE INDEX recommendations_workload_idx ON recommendations(workload_id,priority,status,created_at DESC);
