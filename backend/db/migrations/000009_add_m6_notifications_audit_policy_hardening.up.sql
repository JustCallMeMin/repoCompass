-- Migration: 000009 add M6 notification, audit, and policy hardening.

ALTER TABLE policies
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE policies
    ADD CONSTRAINT policies_status_check CHECK (status IN ('active', 'draft', 'disabled'));

CREATE INDEX IF NOT EXISTS idx_policies_org_status_name
    ON policies (organization_id, status, name);

CREATE TABLE IF NOT EXISTS notification_preferences (
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL,
    channels        JSONB NOT NULL DEFAULT '["in_app"]',
    event_types     JSONB NOT NULL DEFAULT '[]',
    webhook_url     TEXT,
    email           TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         TEXT,
    type            TEXT NOT NULL,
    severity        TEXT NOT NULL DEFAULT 'info',
    title           TEXT NOT NULL,
    message         TEXT NOT NULL,
    resource_type   TEXT,
    resource_id     TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notifications_severity_check CHECK (severity IN ('info', 'warning', 'critical'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_org_user_created
    ON notifications (organization_id, user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_deliveries (
    id              TEXT PRIMARY KEY,
    notification_id TEXT NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    channel         TEXT NOT NULL,
    status          TEXT NOT NULL,
    target          TEXT,
    error_message   TEXT,
    attempted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notification_deliveries_channel_check CHECK (channel IN ('in_app', 'webhook', 'email')),
    CONSTRAINT notification_deliveries_status_check CHECK (status IN ('queued', 'sent', 'failed', 'stubbed'))
);

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notification
    ON notification_deliveries (notification_id);

CREATE TABLE IF NOT EXISTS audit_events (
    id              TEXT PRIMARY KEY,
    type            TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id        TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    details         JSONB NOT NULL DEFAULT '{}',
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_org_time
    ON audit_events (organization_id, occurred_at DESC);
