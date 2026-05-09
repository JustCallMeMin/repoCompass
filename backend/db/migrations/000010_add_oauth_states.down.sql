-- Roll back OAuth state replay protection.

DROP INDEX IF EXISTS idx_oauth_states_expires_at;
DROP TABLE IF EXISTS oauth_states;
