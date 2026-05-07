-- Migration: 000004 harden persistence history indexes and constraints
-- Adds query-path indexes and scan status constraints required by Milestone 3.

CREATE INDEX IF NOT EXISTS idx_repositories_full_name
    ON repositories(full_name);

CREATE INDEX IF NOT EXISTS idx_snapshots_repository_captured_at_desc
    ON repository_snapshots(repository_id, captured_at DESC);

CREATE INDEX IF NOT EXISTS idx_scans_snapshot_time_desc
    ON scans(snapshot_id, COALESCE(end_time, start_time, created_at) DESC);

CREATE INDEX IF NOT EXISTS idx_findings_scan_severity
    ON findings(scan_id, severity);

CREATE INDEX IF NOT EXISTS idx_finding_evidences_finding_id
    ON finding_evidences(finding_id);

CREATE INDEX IF NOT EXISTS idx_finding_evidences_path
    ON finding_evidences(path);

CREATE INDEX IF NOT EXISTS idx_reports_scan_id
    ON reports(scan_id);

CREATE INDEX IF NOT EXISTS idx_reports_scan_format
    ON reports(scan_id, format);

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_metric_created_at
    ON metric_snapshots(metric_key, created_at DESC);

ALTER TABLE scans
    ADD CONSTRAINT scans_status_check
    CHECK (status IN ('created', 'queued', 'running', 'completed', 'failed', 'cancelled'));
