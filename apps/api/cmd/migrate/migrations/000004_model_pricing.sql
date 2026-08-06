CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE provider_models (
    provider TEXT NOT NULL, model TEXT NOT NULL, context_window BIGINT NOT NULL CHECK (context_window > 0),
    supports_structured_output BOOLEAN NOT NULL DEFAULT FALSE, supports_tool_calling BOOLEAN NOT NULL DEFAULT FALSE,
    supports_caching BOOLEAN NOT NULL DEFAULT FALSE,
    deployment_type TEXT NOT NULL CHECK (deployment_type IN ('hosted','local','private_hosted')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','deprecated')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (provider,model)
);
CREATE TABLE model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), provider TEXT NOT NULL, model TEXT NOT NULL,
    input_price NUMERIC(24,12) NOT NULL CHECK (input_price >= 0), output_price NUMERIC(24,12) NOT NULL CHECK (output_price >= 0),
    cached_input_price NUMERIC(24,12) CHECK (cached_input_price >= 0), currency CHAR(3) NOT NULL,
    unit_size BIGINT NOT NULL CHECK (unit_size > 0), effective_from TIMESTAMPTZ NOT NULL, effective_to TIMESTAMPTZ,
    source_note TEXT NOT NULL, illustrative BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (provider,model) REFERENCES provider_models(provider,model),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    EXCLUDE USING gist (provider WITH =, model WITH =, currency WITH =,
        tstzrange(effective_from,COALESCE(effective_to,'infinity'::timestamptz),'[)') WITH &&)
);
CREATE INDEX model_pricing_lookup_idx ON model_pricing(provider,model,currency,effective_from DESC);

INSERT INTO provider_models(provider,model,context_window,supports_structured_output,supports_tool_calling,supports_caching,deployment_type)
VALUES ('fictional-cloud','illustrative-frontier',128000,TRUE,TRUE,TRUE,'hosted'),
       ('fictional-cloud','illustrative-small',32000,TRUE,FALSE,TRUE,'hosted'),
       ('fictional-local','illustrative-local',16000,TRUE,FALSE,FALSE,'local');
INSERT INTO model_pricing(provider,model,input_price,output_price,cached_input_price,currency,unit_size,effective_from,source_note,illustrative)
VALUES ('fictional-cloud','illustrative-frontier',5,15,1,'USD',1000000,'2026-01-01T00:00:00Z','Illustrative portfolio pricing; not a current provider price.',TRUE),
       ('fictional-cloud','illustrative-small',0.5,1.5,0.1,'USD',1000000,'2026-01-01T00:00:00Z','Illustrative portfolio pricing; not a current provider price.',TRUE);
