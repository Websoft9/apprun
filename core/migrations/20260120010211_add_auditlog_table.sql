-- Create "audit_logs" table
CREATE TABLE "audit_logs" (
  "id" uuid NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "operator_id" uuid NULL,
  "action" character varying NOT NULL,
  "target_id" character varying NULL,
  "target_type" character varying NULL,
  "changes" jsonb NULL,
  "ip_address" character varying NULL,
  "user_agent" character varying NULL,
  "status_code" bigint NULL,
  "response_time_ms" bigint NULL,
  "method" character varying NULL,
  "path" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "auditlog_action" to table: "audit_logs"
CREATE INDEX "auditlog_action" ON "audit_logs" ("action");
-- Create index "auditlog_operator_id" to table: "audit_logs"
CREATE INDEX "auditlog_operator_id" ON "audit_logs" ("operator_id");
-- Create index "auditlog_target_type_target_id" to table: "audit_logs"
CREATE INDEX "auditlog_target_type_target_id" ON "audit_logs" ("target_type", "target_id");
-- Create index "auditlog_timestamp" to table: "audit_logs"
CREATE INDEX "auditlog_timestamp" ON "audit_logs" ("timestamp");
