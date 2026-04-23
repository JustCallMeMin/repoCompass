CREATE TABLE IF NOT EXISTS schema_bootstrap_checks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    note TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_bootstrap_checks (note)
VALUES ('bootstrap migration applied');
