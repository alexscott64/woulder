-- Migration 000022 rollback: Remove job monitoring system

DROP TRIGGER IF EXISTS update_job_executions_updated_at ON woulder.job_executions;
DROP INDEX IF EXISTS woulder.idx_job_executions_started_at;
DROP INDEX IF EXISTS woulder.idx_job_executions_active;
DROP INDEX IF EXISTS woulder.idx_job_executions_job_name;
DROP INDEX IF EXISTS woulder.idx_job_executions_status;
DROP TABLE IF EXISTS woulder.job_executions;
