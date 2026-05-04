-- Migration: 000003 create analysis persistence tables
-- Stores rule metadata, analyzer results, findings, evidence, recommendations,
-- assessments, metric snapshots, and report metadata.

CREATE TABLE IF NOT EXISTS rule_sets (
    id         TEXT        NOT NULL PRIMARY KEY,
    name       TEXT        NOT NULL,
    version    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rules (
    id                 TEXT        NOT NULL PRIMARY KEY,
    title              TEXT        NOT NULL,
    description        TEXT        NOT NULL,
    severity           TEXT        NOT NULL,
    category           TEXT        NOT NULL,
    enabled_by_default BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rules_severity_check CHECK (severity IN ('low', 'medium', 'high')),
    CONSTRAINT rules_category_check CHECK (category IN ('documentation', 'workflow', 'ci', 'maintainability'))
);

CREATE TABLE IF NOT EXISTS rule_set_rules (
    rule_set_id TEXT NOT NULL REFERENCES rule_sets(id) ON DELETE CASCADE,
    rule_id     TEXT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    rule_order  INT  NOT NULL,
    PRIMARY KEY (rule_set_id, rule_id)
);

CREATE TABLE IF NOT EXISTS analyzer_results (
    id            BIGSERIAL PRIMARY KEY,
    scan_id       TEXT        NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    analyzer_id   TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    version       TEXT        NOT NULL,
    status        TEXT        NOT NULL,
    duration_ms   BIGINT      NOT NULL DEFAULT 0,
    metadata      JSONB       NOT NULL DEFAULT '{}',
    error_message TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT analyzer_results_status_check CHECK (status IN ('success', 'skipped', 'failed')),
    UNIQUE (scan_id, analyzer_id)
);

CREATE TABLE IF NOT EXISTS findings (
    id          TEXT        NOT NULL PRIMARY KEY,
    scan_id     TEXT        NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    rule_id     TEXT        NOT NULL,
    analyzer_id TEXT        NOT NULL,
    severity    TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    message     TEXT        NOT NULL,
    category    TEXT        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT findings_severity_check CHECK (severity IN ('low', 'medium', 'high')),
    CONSTRAINT findings_category_check CHECK (category IN ('documentation', 'workflow', 'ci', 'maintainability')),
    CONSTRAINT findings_status_check CHECK (status IN ('open'))
);

CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_rule_id ON findings(rule_id);
CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);

CREATE TABLE IF NOT EXISTS finding_evidences (
    id             BIGSERIAL PRIMARY KEY,
    finding_id     TEXT        NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    evidence_order INT         NOT NULL,
    type           TEXT        NOT NULL,
    message        TEXT        NOT NULL,
    path           TEXT        NOT NULL DEFAULT '',
    value          TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT finding_evidences_type_check CHECK (type IN ('file_exists', 'file_missing', 'pattern_match', 'pattern_missing', 'metadata'))
);

CREATE TABLE IF NOT EXISTS recommendations (
    id         BIGSERIAL PRIMARY KEY,
    finding_id TEXT        NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    action     TEXT        NOT NULL,
    rationale  TEXT        NOT NULL,
    priority   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT recommendations_priority_check CHECK (priority IN ('high', 'medium', 'low'))
);

CREATE TABLE IF NOT EXISTS assessments (
    scan_id            TEXT        NOT NULL PRIMARY KEY REFERENCES scans(id) ON DELETE CASCADE,
    overall_score      INT         NOT NULL,
    label              TEXT        NOT NULL,
    finding_count      INT         NOT NULL,
    severity_counts    JSONB       NOT NULL DEFAULT '{}',
    category_scores    JSONB       NOT NULL DEFAULT '{}',
    category_breakdown JSONB       NOT NULL DEFAULT '{}',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT assessments_score_check CHECK (overall_score >= 0 AND overall_score <= 100),
    CONSTRAINT assessments_label_check CHECK (label IN ('excellent', 'good', 'fair', 'poor'))
);

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id         BIGSERIAL PRIMARY KEY,
    scan_id    TEXT        NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    metric_key TEXT        NOT NULL,
    value      NUMERIC     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_scan_metric ON metric_snapshots(scan_id, metric_key);

CREATE TABLE IF NOT EXISTS reports (
    id           BIGSERIAL PRIMARY KEY,
    scan_id      TEXT        NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    format       TEXT        NOT NULL,
    content_type TEXT        NOT NULL,
    metadata     JSONB       NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reports_format_check CHECK (format IN ('markdown', 'json'))
);
