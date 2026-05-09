-- Migration: 000010 add OAuth state replay protection.

CREATE TABLE IF NOT EXISTS oauth_states (
    state       TEXT PRIMARY KEY,
    provider    TEXT        NOT NULL,
    redirect_to TEXT        NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at
    ON oauth_states (expires_at);
