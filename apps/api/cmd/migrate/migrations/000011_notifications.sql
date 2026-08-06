CREATE TABLE notification_preferences (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  in_app boolean NOT NULL DEFAULT true,
  mobile_push boolean NOT NULL DEFAULT false,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE notifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  event_type text NOT NULL CHECK(event_type IN ('budget_threshold_reached','high_priority_recommendation','experiment_guardrail_violation','automatic_rollback','evaluation_completed','experiment_verified')),
  title text NOT NULL,
  body text NOT NULL,
  deep_link text NOT NULL,
  dedupe_key text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','delivered','failed','read')),
  attempts integer NOT NULL DEFAULT 0,
  next_attempt_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  delivered_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id,dedupe_key)
);
CREATE INDEX notifications_delivery_idx ON notifications(status,next_attempt_at) WHERE status IN ('pending','failed');
CREATE INDEX notifications_user_idx ON notifications(user_id,created_at DESC);
CREATE TABLE mobile_push_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash bytea NOT NULL UNIQUE, encrypted_token bytea NOT NULL, created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, revoked_at timestamptz
);

CREATE FUNCTION notify_budget_alert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at)
  SELECT p.organization_id,u.id,'budget_threshold_reached','Budget threshold reached','A project budget threshold requires review.','ai-execution-advisor://open/alerts/'||NEW.id,'budget:'||NEW.id,'delivered',CURRENT_TIMESTAMP
  FROM projects p JOIN users u ON u.organization_id=p.organization_id LEFT JOIN notification_preferences pref ON pref.user_id=u.id
  WHERE p.id=NEW.project_id AND COALESCE(pref.in_app,true);
  RETURN NEW;
END $$;
CREATE TRIGGER budget_alert_notification AFTER INSERT ON budget_alerts FOR EACH ROW EXECUTE FUNCTION notify_budget_alert();

CREATE FUNCTION notify_recommendation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.priority='high' THEN
    INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at)
    SELECT p.organization_id,u.id,'high_priority_recommendation','High-priority recommendation','New evidence-based recommendation available.','ai-execution-advisor://open/recommendations/'||NEW.id,'recommendation:'||NEW.id,'delivered',CURRENT_TIMESTAMP
    FROM workloads w JOIN projects p ON p.id=w.project_id JOIN users u ON u.organization_id=p.organization_id LEFT JOIN notification_preferences pref ON pref.user_id=u.id
    WHERE w.id=NEW.workload_id AND COALESCE(pref.in_app,true);
  END IF; RETURN NEW;
END $$;
CREATE TRIGGER recommendation_notification AFTER INSERT ON recommendations FOR EACH ROW EXECUTE FUNCTION notify_recommendation();

CREATE FUNCTION notify_experiment_state() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE event_name text; notification_title text;
BEGIN
  IF NEW.status='rolled_back' AND OLD.status<>NEW.status THEN event_name:='automatic_rollback'; notification_title:='Experiment rolled back';
  ELSIF NEW.status='verified' AND OLD.status<>NEW.status THEN event_name:='experiment_verified'; notification_title:='Experiment verified';
  ELSE RETURN NEW; END IF;
  INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at)
  SELECT p.organization_id,u.id,event_name,notification_title,'Experiment status changed after guardrail review.','ai-execution-advisor://open/experiments/'||NEW.id,event_name||':'||NEW.id,'delivered',CURRENT_TIMESTAMP
  FROM workloads w JOIN projects p ON p.id=w.project_id JOIN users u ON u.organization_id=p.organization_id LEFT JOIN notification_preferences pref ON pref.user_id=u.id
  WHERE w.id=NEW.workload_id AND COALESCE(pref.in_app,true);
  IF NEW.status='rolled_back' THEN
    INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at)
    SELECT p.organization_id,u.id,'experiment_guardrail_violation','Experiment guardrail violated','An experiment crossed a configured safety guardrail.','ai-execution-advisor://open/experiments/'||NEW.id,'guardrail:'||NEW.id,'delivered',CURRENT_TIMESTAMP
    FROM workloads w JOIN projects p ON p.id=w.project_id JOIN users u ON u.organization_id=p.organization_id LEFT JOIN notification_preferences pref ON pref.user_id=u.id
    WHERE w.id=NEW.workload_id AND COALESCE(pref.in_app,true);
  END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER experiment_state_notification AFTER UPDATE ON experiments FOR EACH ROW EXECUTE FUNCTION notify_experiment_state();

CREATE FUNCTION notify_evaluation_complete() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.status='completed' AND OLD.status<>NEW.status THEN
    INSERT INTO notifications(organization_id,user_id,event_type,title,body,deep_link,dedupe_key,status,delivered_at)
    SELECT p.organization_id,u.id,'evaluation_completed','Evaluation completed','Candidate evaluation results are ready.','ai-execution-advisor://open/evaluations/'||NEW.id,'evaluation:'||NEW.id,'delivered',CURRENT_TIMESTAMP
    FROM evaluation_datasets d JOIN workloads w ON w.id=d.workload_id JOIN projects p ON p.id=w.project_id JOIN users u ON u.organization_id=p.organization_id LEFT JOIN notification_preferences pref ON pref.user_id=u.id
    WHERE d.id=NEW.dataset_id AND COALESCE(pref.in_app,true);
  END IF; RETURN NEW;
END $$;
CREATE TRIGGER evaluation_complete_notification AFTER UPDATE ON evaluation_runs FOR EACH ROW EXECUTE FUNCTION notify_evaluation_complete();
