-- Migration: 000005 rollback – remove M6 performance indexes

DROP INDEX IF EXISTS idx_org_memberships_org_id;
DROP INDEX IF EXISTS idx_org_memberships_user_id;
DROP INDEX IF EXISTS idx_policies_org_id_name;
DROP INDEX IF EXISTS idx_policies_updated_at;
DROP INDEX IF EXISTS idx_repositories_org_id;
DROP INDEX IF EXISTS idx_scans_status_end_time;
DROP INDEX IF EXISTS idx_metric_snapshots_scan_id;
DROP INDEX IF EXISTS idx_metric_snapshots_metric_key;
