-- Seed the local-dev actor into the default Personal organization.
-- This keeps the Docker/local workflow usable until real session auth exists.

INSERT INTO organization_memberships (organization_id, user_id, role, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000000', 'mock_user', 'owner', NOW(), NOW())
ON CONFLICT (organization_id, user_id) DO UPDATE SET
    role = EXCLUDED.role,
    updated_at = EXCLUDED.updated_at;
