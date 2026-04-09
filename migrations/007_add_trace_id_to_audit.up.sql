-- Add trace_id column to audit_logs table
ALTER TABLE audit_logs ADD COLUMN trace_id UUID;
CREATE INDEX idx_audit_logs_trace_id ON audit_logs(trace_id) WHERE trace_id IS NOT NULL;
