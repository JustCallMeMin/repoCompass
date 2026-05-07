-- Migration: 000008 add M4 product API and GitHub integration tables.

ALTER TABLE repositories
    ADD COLUMN IF NOT EXISTS local_path TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS users (
    id          TEXT        NOT NULL PRIMARY KEY,
    github_id   TEXT        NOT NULL UNIQUE,
    login       TEXT        NOT NULL,
    name        TEXT        NOT NULL DEFAULT '',
    avatar_url  TEXT        NOT NULL DEFAULT '',
    email       TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT        NOT NULL PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS github_integrations (
    id              TEXT        NOT NULL PRIMARY KEY,
    organization_id TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    repository_id   TEXT        NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    owner           TEXT        NOT NULL,
    repo            TEXT        NOT NULL,
    clone_url       TEXT        NOT NULL,
    active          BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, owner, repo)
);

CREATE INDEX IF NOT EXISTS idx_github_integrations_org ON github_integrations(organization_id);
CREATE INDEX IF NOT EXISTS idx_github_integrations_repository ON github_integrations(repository_id);

CREATE TABLE IF NOT EXISTS github_webhook_events (
    id              TEXT        NOT NULL PRIMARY KEY,
    delivery_id     TEXT        NOT NULL UNIQUE,
    event_type      TEXT        NOT NULL,
    repository_full_name TEXT   NOT NULL DEFAULT '',
    repository_clone_url TEXT   NOT NULL DEFAULT '',
    payload         JSONB       NOT NULL DEFAULT '{}',
    status          TEXT        NOT NULL,
    error_message   TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT github_webhook_events_status_check CHECK (status IN ('accepted', 'ignored', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_github_webhook_events_repo ON github_webhook_events(repository_full_name);

CREATE TABLE IF NOT EXISTS scan_jobs (
    id              TEXT        NOT NULL PRIMARY KEY,
    repository_id   TEXT        NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    source_type     TEXT        NOT NULL,
    source_url      TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL,
    scan_id         TEXT        REFERENCES scans(id) ON DELETE SET NULL,
    webhook_event_id TEXT       REFERENCES github_webhook_events(id) ON DELETE SET NULL,
    error_message   TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    CONSTRAINT scan_jobs_source_type_check CHECK (source_type IN ('local', 'github')),
    CONSTRAINT scan_jobs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_scan_jobs_status_created ON scan_jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_repository ON scan_jobs(repository_id);
