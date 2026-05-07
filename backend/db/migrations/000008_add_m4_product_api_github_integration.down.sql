-- Migration: 000008 add M4 product API and GitHub integration tables down.

DROP INDEX IF EXISTS idx_scan_jobs_repository;
DROP INDEX IF EXISTS idx_scan_jobs_status_created;
DROP TABLE IF EXISTS scan_jobs;

DROP INDEX IF EXISTS idx_github_webhook_events_repo;
DROP TABLE IF EXISTS github_webhook_events;

DROP INDEX IF EXISTS idx_github_integrations_repository;
DROP INDEX IF EXISTS idx_github_integrations_org;
DROP TABLE IF EXISTS github_integrations;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;

DROP TABLE IF EXISTS users;

ALTER TABLE repositories
    DROP COLUMN IF EXISTS local_path;
