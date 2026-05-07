-- Migration: 000004 add organization management
-- Creates organizations, organization_memberships, organization_configurations, and policies tables.
-- Modifies existing tables to belong to organizations.

CREATE TABLE IF NOT EXISTS organizations (
    id         TEXT        NOT NULL PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organization_memberships (
    organization_id TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         TEXT        NOT NULL,
    role            TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT memberships_role_check CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

CREATE INDEX IF NOT EXISTS idx_memberships_user_id
    ON organization_memberships(user_id);

CREATE TABLE IF NOT EXISTS organization_configurations (
    organization_id TEXT        NOT NULL PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    settings        JSONB       NOT NULL DEFAULT '{}',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS policies (
    id              TEXT        NOT NULL PRIMARY KEY,
    organization_id TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    rules           JSONB       NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policies_organization_id
    ON policies(organization_id);

-- Note: The `repositories` table already has an `organization_id TEXT NOT NULL DEFAULT ''` column
-- from the initial M5 setup. In this migration, we will backfill the data, and we will
-- add a foreign key constraint to it so it properly references the organizations table.

-- 1. Create a default "Personal" organization to hold existing repositories.
INSERT INTO organizations (id, name, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000000', 'Personal', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 2. Update existing repositories to belong to the default organization if they don't have one.
UPDATE repositories
SET organization_id = '00000000-0000-0000-0000-000000000000'
WHERE organization_id = '';

-- 3. Now we can safely add the foreign key constraint.
-- Because SQLite (if used for testing) doesn't support ADD CONSTRAINT, we only do this for Postgres.
-- Since we are strictly using Postgres as per requirements, this is safe.
ALTER TABLE repositories
ADD CONSTRAINT fk_repositories_organization
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT;
