-- Migration: 000004 add organization management (Down)

-- 1. Drop the foreign key from repositories
ALTER TABLE repositories
DROP CONSTRAINT IF EXISTS fk_repositories_organization;

-- Note: We don't revert the organization_id back to '' because the column exists from M5
-- and having UUIDs in it doesn't break M5's schema constraints, just orphans them visually
-- if we roll back.

-- 2. Drop the policies table
DROP TABLE IF EXISTS policies;

-- 3. Drop organization_configurations
DROP TABLE IF EXISTS organization_configurations;

-- 4. Drop organization_memberships
DROP TABLE IF EXISTS organization_memberships;

-- 5. Drop organizations
DROP TABLE IF EXISTS organizations;
