-- Roll back M6 notification, audit, and policy hardening.

DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_preferences;

DROP INDEX IF EXISTS idx_policies_org_status_name;

ALTER TABLE policies
    DROP CONSTRAINT IF EXISTS policies_status_check;

ALTER TABLE policies
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS version;
