-- Migration: 000002 create core scan tables
-- Creates repositories, repository_snapshots, and scans tables.

CREATE TABLE IF NOT EXISTS repositories (
    id                TEXT        NOT NULL PRIMARY KEY,
    name              TEXT        NOT NULL,
    owner_name        TEXT        NOT NULL DEFAULT '',
    full_name         TEXT        NOT NULL,
    url               TEXT        NOT NULL DEFAULT '',
    provider          TEXT        NOT NULL,
    default_branch    TEXT        NOT NULL DEFAULT '',
    primary_ecosystem TEXT        NOT NULL DEFAULT '',
    is_monorepo       BOOLEAN     NOT NULL DEFAULT FALSE,
    status            TEXT        NOT NULL DEFAULT 'active',
    organization_id   TEXT        NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS repository_snapshots (
    id                TEXT        NOT NULL PRIMARY KEY,
    repository_id     TEXT        NOT NULL REFERENCES repositories(id),
    source_type       TEXT        NOT NULL,
    branch_name       TEXT        NOT NULL DEFAULT '',
    commit_sha        TEXT        NOT NULL DEFAULT '',
    tree_reference    TEXT        NOT NULL DEFAULT '',
    captured_at       TIMESTAMPTZ NOT NULL,
    snapshot_metadata JSONB       NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_snapshots_repository_id
    ON repository_snapshots(repository_id);

CREATE TABLE IF NOT EXISTS scans (
    id            TEXT        NOT NULL PRIMARY KEY,
    snapshot_id   TEXT        NOT NULL REFERENCES repository_snapshots(id),
    status        TEXT        NOT NULL,
    start_time    TIMESTAMPTZ,
    end_time      TIMESTAMPTZ,
    error_details TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scans_snapshot_id
    ON scans(snapshot_id);

CREATE INDEX IF NOT EXISTS idx_scans_status
    ON scans(status);
