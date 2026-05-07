-- Rollback: 000007 harden persistence history indexes and constraints

ALTER TABLE scans
    DROP CONSTRAINT IF EXISTS scans_status_check;

DROP INDEX IF EXISTS idx_metric_snapshots_metric_created_at;
DROP INDEX IF EXISTS idx_reports_scan_format;
DROP INDEX IF EXISTS idx_reports_scan_id;
DROP INDEX IF EXISTS idx_finding_evidences_path;
DROP INDEX IF EXISTS idx_finding_evidences_finding_id;
DROP INDEX IF EXISTS idx_findings_scan_severity;
DROP INDEX IF EXISTS idx_scans_snapshot_time_desc;
DROP INDEX IF EXISTS idx_snapshots_repository_captured_at_desc;
DROP INDEX IF EXISTS idx_repositories_full_name;
