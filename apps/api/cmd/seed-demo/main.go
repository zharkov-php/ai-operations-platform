package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
)

func main() {
	if err := run(context.Background(), os.Getenv("DATABASE_URL")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("deterministic demo data seeded")
}
func run(ctx context.Context, databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, seedSQL)
	return err
}

const seedSQL = `
DO $$ DECLARE
 org uuid; project uuid; current_workload uuid; i int;
 names text[]:=ARRAY['VAT Calculator','Internal Email Translator','Ticket Classifier','Architecture Review','Duplicate FAQ Generator'];
 slugs text[]:=ARRAY['vat-calculator','internal-email-translator','ticket-classifier','architecture-review','duplicate-faq-generator'];
 task_types text[]:=ARRAY['calculation','translation','classification','reasoning','generation'];
 rec_types text[]:=ARRAY['replace_with_code','local_model','smaller_hosted_model','keep_current_model','exact_cache'];
BEGIN
 SELECT id INTO org FROM organizations WHERE slug='demo';
 IF org IS NULL THEN RAISE EXCEPTION 'run seed-auth first'; END IF;
 INSERT INTO projects(organization_id,name,slug,environment,monthly_budget,currency) VALUES(org,'AI Product','ai-product','production',250,'USD') ON CONFLICT(organization_id,slug) DO UPDATE SET monthly_budget=EXCLUDED.monthly_budget RETURNING id INTO project;
 FOR i IN 1..5 LOOP
  INSERT INTO workloads(project_id,name,slug,type,owner,criticality,quality_requirement,privacy_classification)
  VALUES(project,names[i],slugs[i],CASE WHEN i=4 THEN 'internal_tool' ELSE 'feature' END,'Demo Team',CASE WHEN i=4 THEN 'critical' ELSE 'medium' END,CASE WHEN i=4 THEN 'critical' ELSE 'standard' END,CASE WHEN i=2 THEN 'confidential' ELSE 'internal' END)
  ON CONFLICT(project_id,slug) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO current_workload;
  INSERT INTO task_classifications(workload_id,task_type,determinism,complexity,privacy_classification,quality_requirement,failure_impact,external_knowledge_required,natural_language_understanding_required,confidence,classification_source,evidence)
  VALUES(current_workload,task_types[i],CASE WHEN i=1 THEN 'high' ELSE 'low' END,CASE WHEN i=4 THEN 'high' ELSE 'low' END,CASE WHEN i=2 THEN 'confidential' ELSE 'internal' END,CASE WHEN i=4 THEN 'critical' ELSE 'standard' END,CASE WHEN i=4 THEN 'high' ELSE 'low' END,false,i<>1,0.9,'demo-seed','{"source":"deterministic demo"}') ON CONFLICT(workload_id,classification_source) DO NOTHING;
  INSERT INTO recommendations(workload_id,recommendation_type,priority,confidence_level,confidence,current_execution,proposed_execution,reason_codes,evidence_summary,confidence_inputs,estimated_monthly_savings,currency,estimated_implementation_cost,estimated_break_even_months,quality_risk,operational_risk,required_next_action,rule_version)
  VALUES(current_workload,rec_types[i],CASE WHEN i IN(1,3) THEN 'high' ELSE 'medium' END,CASE WHEN i=4 THEN 'medium' ELSE 'high' END,0.85,'{"provider":"fictional-cloud","model":"illustrative-frontier"}',jsonb_build_object('candidate',rec_types[i]),ARRAY['demo_evidence'],jsonb_build_object('summary','Deterministic seeded evidence for the portfolio journey.'),'{"analyzed_calls":150}',CASE WHEN i=4 THEN 0 ELSE 35 END,'USD',CASE WHEN i=1 THEN 200 ELSE 100 END,CASE WHEN i=4 THEN NULL ELSE 5.7 END,CASE WHEN i=4 THEN 'high' ELSE 'medium' END,'low',CASE WHEN i=4 THEN 'retain current execution' ELSE 'evaluation required' END,'demo-v1') ON CONFLICT(workload_id,recommendation_type,rule_version) DO UPDATE SET updated_at=CURRENT_TIMESTAMP;
  INSERT INTO llm_calls(organization_id,project_id,workload_id,external_call_id,trace_id,provider,model,prompt_template_id,request_timestamp,response_timestamp,latency_ms,input_tokens,output_tokens,cached_input_tokens,retry_count,status,estimated_cost,currency,prompt_hash,redacted_prompt_preview,redacted_response_preview,metadata,payload_hash)
  SELECT org,project,current_workload,'demo-'||i||'-'||n,'demo-trace-'||n,'fictional-cloud','illustrative-frontier','template-'||i,CURRENT_TIMESTAMP-(n||' hours')::interval,CURRENT_TIMESTAMP-(n||' hours')::interval+interval '1 second',500+i*100,800*i,120*i,0,CASE WHEN i=3 AND n%10=0 THEN 1 ELSE 0 END,'success',(800*i*5+120*i*15)::numeric/1000000,'USD',digest(CASE WHEN i=5 THEN 'duplicate' ELSE 'prompt-'||n END,'sha256'),'[REDACTED]','[REDACTED]',jsonb_build_object('demo',true),digest('payload-'||i||'-'||n,'sha256') FROM generate_series(1,150) n ON CONFLICT(organization_id,external_call_id) DO NOTHING;
 END LOOP;
 INSERT INTO project_budget_thresholds(project_id,percentage,severity) VALUES(project,50,'warning') ON CONFLICT(project_id,percentage) DO NOTHING;
 INSERT INTO budget_alerts(project_id,threshold_type,threshold_value,severity,dedupe_key,evidence) VALUES(project,'projected_overspend',275,'warning','demo-projected','{"projection":"275","budget":"250"}') ON CONFLICT(project_id,dedupe_key) WHERE status='open' DO NOTHING;
 INSERT INTO experiments(workload_id,recommendation_id,control_execution,candidate_execution,traffic_percentage,status,guardrails,started_at,completed_at,results,verified_savings,currency)
 SELECT r.workload_id,r.id,'{"model":"illustrative-frontier"}','{"model":"illustrative-small"}',20,'verified','{"minimum_quality_score":"0.90","maximum_latency_ms":"1000","maximum_error_rate":"0.02","maximum_cost_per_call":"0.01"}',CURRENT_TIMESTAMP-interval '7 days',CURRENT_TIMESTAMP-interval '1 day','{"quality_score":"0.95"}',18.5,'USD' FROM recommendations r JOIN workloads w ON w.id=r.workload_id WHERE w.slug='ticket-classifier' AND r.recommendation_type='smaller_hosted_model' AND NOT EXISTS(SELECT 1 FROM experiments e WHERE e.recommendation_id=r.id);
 INSERT INTO experiments(workload_id,recommendation_id,control_execution,candidate_execution,traffic_percentage,status,guardrails,started_at,completed_at,rollback_reason,results,currency)
 SELECT r.workload_id,r.id,'{"model":"illustrative-frontier"}','{"model":"illustrative-local"}',10,'rolled_back','{"minimum_quality_score":"0.90","maximum_latency_ms":"1000","maximum_error_rate":"0.02","maximum_cost_per_call":"0.01"}',CURRENT_TIMESTAMP-interval '2 days',CURRENT_TIMESTAMP-interval '1 day','Quality guardrail violated','{"quality_score":"0.72"}','USD' FROM recommendations r JOIN workloads w ON w.id=r.workload_id WHERE w.slug='internal-email-translator' AND r.recommendation_type='local_model' AND NOT EXISTS(SELECT 1 FROM experiments e WHERE e.recommendation_id=r.id);
END $$;`
