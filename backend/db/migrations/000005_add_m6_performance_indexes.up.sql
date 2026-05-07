-- Migration: 000005 add M6 performance indexes and pagination support
-- Adds indexes on org-scoped query hot paths and ensures
-- all list queries support cursor/limit-based pagination.

-- Org membership lookups
CREATE INDEX IF NOT EXISTS idx_org_memberships_org_id
    ON organization_memberships (organization_id);

CREATE INDEX IF NOT EXISTS idx_org_memberships_user_id
    ON organization_memberships (user_id);

-- Policies by org + name (used for GetActiveAssessmentPolicy)
CREATE INDEX IF NOT EXISTS idx_policies_org_id_name
    ON policies (organization_id, name);

CREATE INDEX IF NOT EXISTS idx_policies_updated_at
    ON policies (updated_at DESC);

-- Repositories by org (used for org repos view and insights)
CREATE INDEX IF NOT EXISTS idx_repositories_org_id
    ON repositories (organization_id);

-- Scans by status (used in insights aggregate and benchmark queries)
CREATE INDEX IF NOT EXISTS idx_scans_status_end_time
    ON scans (status, end_time DESC);

-- Metric snapshots for trend queries
CREATE INDEX IF NOT EXISTS idx_metric_snapshots_scan_id
    ON metric_snapshots (scan_id);

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_metric_key
    ON metric_snapshots (metric_key);
