// Package postgres provides a PostgreSQL implementation of scan.Store.
// It uses database/sql with the pgx/v5 driver via github.com/jackc/pgx/v5/stdlib.
//
// Usage:
//
//	db, err := postgres.Open(databaseURL)
//	if err != nil { ... }
//	defer db.Close()
//	store := postgres.New(db)
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/history"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Store is a PostgreSQL-backed implementation of scan.Store.
type Store struct {
	db *sql.DB
}

// Open opens a connection pool to the Postgres database at dsn and pings it.
// The caller is responsible for calling db.Close() when done.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, databaseError("open connection pool", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, databaseError("ping database", err)
	}
	return db, nil
}

// New creates a Store backed by the provided *sql.DB.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// SaveRepository inserts or updates a repository.
func (s *Store) SaveRepository(ctx context.Context, repo repository.Repository) error {
	if err := validateRepository(repo); err != nil {
		return err
	}
	orgID := repo.OrganizationID
	if orgID == "" {
		orgID = org.DefaultPersonalOrgID
	}
	const q = `
INSERT INTO repositories (
	id, name, owner_name, full_name, url, local_path, provider,
	default_branch, primary_ecosystem, is_monorepo, status,
	organization_id, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7,
	$8, $9, $10, $11,
	$12, NOW()
)
ON CONFLICT (id) DO UPDATE SET
	name              = EXCLUDED.name,
	owner_name        = EXCLUDED.owner_name,
	full_name         = EXCLUDED.full_name,
	url               = EXCLUDED.url,
	local_path        = EXCLUDED.local_path,
	provider          = EXCLUDED.provider,
	default_branch    = EXCLUDED.default_branch,
	primary_ecosystem = EXCLUDED.primary_ecosystem,
	is_monorepo       = EXCLUDED.is_monorepo,
	status            = EXCLUDED.status,
	organization_id   = EXCLUDED.organization_id,
	updated_at        = NOW()`

	_, err := s.db.ExecContext(ctx, q,
		repo.ID,
		repo.Name,
		repo.OwnerName,
		repo.FullName,
		repo.URL,
		repo.LocalPath,
		string(repo.Provider),
		repo.DefaultBranch,
		repo.PrimaryEcosystem,
		repo.IsMonorepo,
		string(repo.Status),
		orgID,
	)
	if err != nil {
		return databaseError("save repository", err)
	}
	return nil
}

// SaveSnapshot inserts a new repository_snapshot row.
func (s *Store) SaveSnapshot(ctx context.Context, snap snapshot.RepositorySnapshot) error {
	if err := validateSnapshot(snap); err != nil {
		return err
	}
	metadata, err := marshalMetadata(snap.SnapshotMetadata)
	if err != nil {
		return databaseError("marshal snapshot metadata", err)
	}

	const q = `
INSERT INTO repository_snapshots (
	id, repository_id, source_type, branch_name, commit_sha,
	tree_reference, captured_at, snapshot_metadata
) VALUES (
	$1, $2, $3, $4, $5,
	$6, $7, $8
)`

	_, err = s.db.ExecContext(ctx, q,
		snap.ID,
		snap.RepositoryID,
		string(snap.SourceType),
		snap.BranchName,
		snap.CommitSHA,
		snap.TreeReference,
		snap.CapturedAt,
		metadata,
	)
	if err != nil {
		return databaseError("save snapshot", err)
	}
	return nil
}

// SaveScan inserts a new scan row with its initial state.
func (s *Store) SaveScan(ctx context.Context, sc scan.Scan) error {
	if err := validateScan(sc); err != nil {
		return err
	}
	const q = `
INSERT INTO scans (
	id, snapshot_id, status, start_time, end_time, error_details
) VALUES (
	$1, $2, $3, $4, $5, $6
)`

	_, err := s.db.ExecContext(ctx, q,
		sc.ID,
		sc.SnapshotID,
		string(sc.Status),
		sc.StartTime,
		sc.EndTime,
		sc.ErrorDetails,
	)
	if err != nil {
		return databaseError("save scan", err)
	}
	return nil
}

// UpdateScan updates the mutable fields of an existing scan row.
func (s *Store) UpdateScan(ctx context.Context, sc scan.Scan) error {
	const q = `
UPDATE scans
SET status        = $1,
    end_time      = $2,
    error_details = $3
WHERE id = $4`

	_, err := s.db.ExecContext(ctx, q,
		string(sc.Status),
		sc.EndTime,
		sc.ErrorDetails,
		sc.ID,
	)
	if err != nil {
		return databaseError("update scan", err)
	}
	return nil
}

// SaveRunResult persists a completed scan result in a single transaction.
func (s *Store) SaveRunResult(ctx context.Context, result scan.RunResult) error {
	if err := validateRunResult(result); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return databaseError("begin scan result transaction", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := saveRepository(ctx, tx, result.Repository); err != nil {
		return err
	}
	if err := saveSnapshot(ctx, tx, result.Snapshot); err != nil {
		return err
	}
	if err := saveScan(ctx, tx, result.Scan); err != nil {
		return err
	}
	if err := saveDefaultRuleSet(ctx, tx); err != nil {
		return err
	}
	if err := saveAnalyzerResults(ctx, tx, result.Scan.ID, result.AnalyzerResults); err != nil {
		return err
	}
	allFindings := collectFindings(result.AnalyzerResults)
	if err := saveFindings(ctx, tx, result.Scan.ID, allFindings); err != nil {
		return err
	}
	if err := saveAssessment(ctx, tx, result.Scan.ID, result.Assessment); err != nil {
		return err
	}
	if err := saveMetricSnapshots(ctx, tx, result.Scan.ID, result.Assessment); err != nil {
		return err
	}
	if err := saveReportMetadata(ctx, tx, result.Scan.ID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return databaseError("commit scan result transaction", err)
	}
	return nil
}

// ListScanHistory returns persisted scans for one repository, newest first.
func (s *Store) ListScanHistory(ctx context.Context, repositoryID string, limit int) ([]history.ScanSummary, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT s.id, rs.repository_id, s.snapshot_id, s.status, s.start_time, s.end_time,
       COALESCE(a.overall_score, 0), COALESCE(a.label, ''), COALESCE(a.finding_count, 0)
FROM scans s
JOIN repository_snapshots rs ON rs.id = s.snapshot_id
LEFT JOIN assessments a ON a.scan_id = s.id
WHERE rs.repository_id = $1
ORDER BY COALESCE(s.end_time, s.start_time, s.created_at) DESC
LIMIT $2`, repositoryID, limit)
	if err != nil {
		return nil, databaseError("list scan history", err)
	}
	defer rows.Close()

	var summaries []history.ScanSummary
	for rows.Next() {
		var item history.ScanSummary
		if err := rows.Scan(
			&item.ScanID,
			&item.RepositoryID,
			&item.SnapshotID,
			&item.Status,
			&item.StartedAt,
			&item.CompletedAt,
			&item.Score,
			&item.Label,
			&item.FindingCount,
		); err != nil {
			return nil, databaseError("scan history row", err)
		}
		summaries = append(summaries, item)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseError("scan history rows", err)
	}
	return summaries, nil
}

// LatestScan returns the newest persisted scan for one repository.
func (s *Store) LatestScan(ctx context.Context, repositoryID string) (history.ScanSummary, error) {
	items, err := s.ListScanHistory(ctx, repositoryID, 1)
	if err != nil {
		return history.ScanSummary{}, err
	}
	if len(items) == 0 {
		return history.ScanSummary{}, databaseError("latest scan", sql.ErrNoRows)
	}
	return items[0], nil
}

// LatestSnapshot returns the newest persisted snapshot for one repository.
func (s *Store) LatestSnapshot(ctx context.Context, repositoryID string) (history.SnapshotSummary, error) {
	var item history.SnapshotSummary
	err := s.db.QueryRowContext(ctx, `
SELECT id, repository_id, source_type, branch_name, commit_sha, tree_reference, captured_at
FROM repository_snapshots
WHERE repository_id = $1
ORDER BY captured_at DESC
LIMIT 1`, repositoryID).Scan(
		&item.SnapshotID,
		&item.RepositoryID,
		&item.SourceType,
		&item.BranchName,
		&item.CommitSHA,
		&item.TreeReference,
		&item.CapturedAt,
	)
	if err != nil {
		return history.SnapshotSummary{}, databaseError("latest snapshot", err)
	}
	return item, nil
}

// ListFindings returns findings with evidence and recommendations for one scan.
func (s *Store) ListFindings(ctx context.Context, scanID string) ([]history.FindingDetail, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, scan_id, rule_id, analyzer_id, severity, title, message, category, status
FROM findings
WHERE scan_id = $1
ORDER BY severity DESC, id ASC`, scanID)
	if err != nil {
		return nil, databaseError("list findings", err)
	}
	defer rows.Close()

	var items []history.FindingDetail
	for rows.Next() {
		var item history.FindingDetail
		if err := rows.Scan(
			&item.ID,
			&item.ScanID,
			&item.RuleID,
			&item.AnalyzerID,
			&item.Severity,
			&item.Title,
			&item.Message,
			&item.Category,
			&item.Status,
		); err != nil {
			return nil, databaseError("finding row", err)
		}
		evidence, err := s.listEvidence(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		recommendations, err := s.listRecommendations(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Evidence = evidence
		item.Recommendations = recommendations
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseError("finding rows", err)
	}
	return items, nil
}

// ListMetricTrend returns one metric series for one repository, oldest first.
func (s *Store) ListMetricTrend(ctx context.Context, repositoryID string, metricKey string, limit int) ([]history.MetricPoint, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT ms.scan_id, ms.metric_key, ms.value::float8, ms.created_at, sc.end_time, rs.repository_id
FROM metric_snapshots ms
JOIN scans sc ON sc.id = ms.scan_id
JOIN repository_snapshots rs ON rs.id = sc.snapshot_id
WHERE rs.repository_id = $1 AND ms.metric_key = $2
ORDER BY COALESCE(sc.end_time, sc.start_time, sc.created_at) ASC
LIMIT $3`, repositoryID, metricKey, limit)
	if err != nil {
		return nil, databaseError("list metric trend", err)
	}
	defer rows.Close()

	var points []history.MetricPoint
	for rows.Next() {
		var point history.MetricPoint
		if err := rows.Scan(
			&point.ScanID,
			&point.MetricKey,
			&point.Value,
			&point.CapturedAt,
			&point.CompletedAt,
			&point.RepositoryID,
		); err != nil {
			return nil, databaseError("metric trend row", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, databaseError("metric trend rows", err)
	}
	return points, nil
}

// marshalMetadata serialises a metadata map to a JSON byte slice suitable for
// storing in a JSONB column. A nil map marshals to the JSON object "{}".
func marshalMetadata(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// saveRepository is the internal transaction-aware version.
func saveRepository(ctx context.Context, exec dbExecutor, repo repository.Repository) error {
	if err := validateRepository(repo); err != nil {
		return err
	}
	orgID := repo.OrganizationID
	if orgID == "" {
		orgID = org.DefaultPersonalOrgID
	}
	const q = `
INSERT INTO repositories (
	id, name, owner_name, full_name, url, local_path, provider,
	default_branch, primary_ecosystem, is_monorepo, status,
	organization_id, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7,
	$8, $9, $10, $11,
	$12, NOW()
)
ON CONFLICT (id) DO UPDATE SET
	name              = EXCLUDED.name,
	owner_name        = EXCLUDED.owner_name,
	full_name         = EXCLUDED.full_name,
	url               = EXCLUDED.url,
	local_path        = EXCLUDED.local_path,
	provider          = EXCLUDED.provider,
	default_branch    = EXCLUDED.default_branch,
	primary_ecosystem = EXCLUDED.primary_ecosystem,
	is_monorepo       = EXCLUDED.is_monorepo,
	status            = EXCLUDED.status,
	organization_id   = EXCLUDED.organization_id,
	updated_at        = NOW()`
	_, err := exec.ExecContext(ctx, q,
		repo.ID,
		repo.Name,
		repo.OwnerName,
		repo.FullName,
		repo.URL,
		repo.LocalPath,
		string(repo.Provider),
		repo.DefaultBranch,
		repo.PrimaryEcosystem,
		repo.IsMonorepo,
		string(repo.Status),
		orgID,
	)
	if err != nil {
		return fmt.Errorf("postgres: save repository: %w", err)
	}
	return nil
}

// listEvidence returns stored evidence rows for one finding.
func (s *Store) listEvidence(ctx context.Context, findingID string) ([]history.EvidenceDetail, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT type, message, path, value
FROM finding_evidences
WHERE finding_id = $1
ORDER BY evidence_order ASC, id ASC`, findingID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list evidence for finding %q: %w", findingID, err)
	}
	defer rows.Close()

	var items []history.EvidenceDetail
	for rows.Next() {
		var item history.EvidenceDetail
		if err := rows.Scan(&item.Type, &item.Message, &item.Path, &item.Value); err != nil {
			return nil, fmt.Errorf("postgres: list evidence for finding %q: scan row: %w", findingID, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list evidence for finding %q: rows: %w", findingID, err)
	}
	return items, nil
}

// listRecommendations returns stored recommendation rows for one finding.
func (s *Store) listRecommendations(ctx context.Context, findingID string) ([]history.RecommendationDetail, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT title, action, rationale, priority
FROM recommendations
WHERE finding_id = $1
ORDER BY id ASC`, findingID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list recommendations for finding %q: %w", findingID, err)
	}
	defer rows.Close()

	var items []history.RecommendationDetail
	for rows.Next() {
		var item history.RecommendationDetail
		if err := rows.Scan(&item.Title, &item.Action, &item.Rationale, &item.Priority); err != nil {
			return nil, fmt.Errorf("postgres: list recommendations for finding %q: scan row: %w", findingID, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list recommendations for finding %q: rows: %w", findingID, err)
	}
	return items, nil
}

// saveSnapshot upserts a repository snapshot using the provided executor.
func saveSnapshot(ctx context.Context, exec dbExecutor, snap snapshot.RepositorySnapshot) error {
	if err := validateSnapshot(snap); err != nil {
		return err
	}
	metadata, err := marshalMetadata(snap.SnapshotMetadata)
	if err != nil {
		return fmt.Errorf("postgres: save snapshot: marshal metadata: %w", err)
	}
	const q = `
INSERT INTO repository_snapshots (
	id, repository_id, source_type, branch_name, commit_sha,
	tree_reference, captured_at, snapshot_metadata
) VALUES (
	$1, $2, $3, $4, $5,
	$6, $7, $8
)
ON CONFLICT (id) DO UPDATE SET
	repository_id     = EXCLUDED.repository_id,
	source_type       = EXCLUDED.source_type,
	branch_name       = EXCLUDED.branch_name,
	commit_sha        = EXCLUDED.commit_sha,
	tree_reference    = EXCLUDED.tree_reference,
	captured_at       = EXCLUDED.captured_at,
	snapshot_metadata = EXCLUDED.snapshot_metadata`
	_, err = exec.ExecContext(ctx, q,
		snap.ID,
		snap.RepositoryID,
		string(snap.SourceType),
		snap.BranchName,
		snap.CommitSHA,
		snap.TreeReference,
		snap.CapturedAt,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("postgres: save snapshot: %w", err)
	}
	return nil
}

// saveScan upserts one scan row using the provided executor.
func saveScan(ctx context.Context, exec dbExecutor, sc scan.Scan) error {
	if err := validateScan(sc); err != nil {
		return err
	}
	const q = `
INSERT INTO scans (
	id, snapshot_id, status, start_time, end_time, error_details
) VALUES (
	$1, $2, $3, $4, $5, $6
)
ON CONFLICT (id) DO UPDATE SET
	snapshot_id    = EXCLUDED.snapshot_id,
	status         = EXCLUDED.status,
	start_time     = EXCLUDED.start_time,
	end_time       = EXCLUDED.end_time,
	error_details  = EXCLUDED.error_details`
	_, err := exec.ExecContext(ctx, q,
		sc.ID,
		sc.SnapshotID,
		string(sc.Status),
		sc.StartTime,
		sc.EndTime,
		sc.ErrorDetails,
	)
	if err != nil {
		return fmt.Errorf("postgres: save scan: %w", err)
	}
	return nil
}

// saveDefaultRuleSet upserts the built-in rules and ruleset metadata.
func saveDefaultRuleSet(ctx context.Context, exec dbExecutor) error {
	ruleSet := rules.DefaultRuleSet()
	if _, err := exec.ExecContext(ctx, `
INSERT INTO rule_sets (id, name, version)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, version = EXCLUDED.version`,
		ruleSet.ID, ruleSet.Name, ruleSet.Version); err != nil {
		return fmt.Errorf("postgres: save ruleset: %w", err)
	}
	for i, rule := range ruleSet.Rules {
		if _, err := exec.ExecContext(ctx, `
INSERT INTO rules (id, title, description, severity, category, enabled_by_default)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	severity = EXCLUDED.severity,
	category = EXCLUDED.category,
	enabled_by_default = EXCLUDED.enabled_by_default`,
			rule.ID, rule.Title, rule.Description, string(rule.Severity), string(rule.Category), rule.EnabledByDefault); err != nil {
			return fmt.Errorf("postgres: save rule %q: %w", rule.ID, err)
		}
		if _, err := exec.ExecContext(ctx, `
INSERT INTO rule_set_rules (rule_set_id, rule_id, rule_order)
VALUES ($1, $2, $3)
ON CONFLICT (rule_set_id, rule_id) DO UPDATE SET rule_order = EXCLUDED.rule_order`,
			ruleSet.ID, rule.ID, i); err != nil {
			return fmt.Errorf("postgres: save ruleset rule %q: %w", rule.ID, err)
		}
	}
	return nil
}

// saveAnalyzerResults upserts analyzer result rows for one scan.
func saveAnalyzerResults(ctx context.Context, exec dbExecutor, scanID string, results []analyzer.AnalyzerResult) error {
	for _, result := range results {
		metadata, err := marshalJSON(result.Metadata)
		if err != nil {
			return fmt.Errorf("postgres: save analyzer result %q: marshal metadata: %w", result.AnalyzerID, err)
		}
		if _, err := exec.ExecContext(ctx, `
INSERT INTO analyzer_results (
	scan_id, analyzer_id, name, version, status, duration_ms, metadata, error_message
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (scan_id, analyzer_id) DO UPDATE SET
	name = EXCLUDED.name,
	version = EXCLUDED.version,
	status = EXCLUDED.status,
	duration_ms = EXCLUDED.duration_ms,
	metadata = EXCLUDED.metadata,
	error_message = EXCLUDED.error_message`,
			scanID,
			result.AnalyzerID,
			result.Name,
			result.Version,
			string(result.Status),
			result.Duration.Milliseconds(),
			metadata,
			result.ErrorMessage,
		); err != nil {
			return fmt.Errorf("postgres: save analyzer result %q: %w", result.AnalyzerID, err)
		}
	}
	return nil
}

// saveFindings upserts findings and replaces child evidence/recommendation rows.
func saveFindings(ctx context.Context, exec dbExecutor, scanID string, values []findings.Finding) error {
	for _, finding := range values {
		if err := validateFinding(finding); err != nil {
			return err
		}
		if _, err := exec.ExecContext(ctx, `
INSERT INTO findings (
	id, scan_id, rule_id, analyzer_id, severity, title, message, category, status
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id) DO UPDATE SET
	scan_id = EXCLUDED.scan_id,
	rule_id = EXCLUDED.rule_id,
	analyzer_id = EXCLUDED.analyzer_id,
	severity = EXCLUDED.severity,
	title = EXCLUDED.title,
	message = EXCLUDED.message,
	category = EXCLUDED.category,
	status = EXCLUDED.status`,
			finding.ID,
			scanID,
			finding.RuleID,
			finding.AnalyzerID,
			string(finding.Severity),
			finding.Title,
			finding.Message,
			string(finding.Category),
			string(finding.Status),
		); err != nil {
			return fmt.Errorf("postgres: save finding %q: %w", finding.ID, err)
		}
		if _, err := exec.ExecContext(ctx, `DELETE FROM finding_evidences WHERE finding_id = $1`, finding.ID); err != nil {
			return fmt.Errorf("postgres: replace evidence for finding %q: %w", finding.ID, err)
		}
		for i, evidence := range finding.Evidence {
			if _, err := exec.ExecContext(ctx, `
INSERT INTO finding_evidences (finding_id, evidence_order, type, message, path, value)
VALUES ($1, $2, $3, $4, $5, $6)`,
				finding.ID, i, string(evidence.Type), evidence.Message, evidence.Path, evidence.Value); err != nil {
				return fmt.Errorf("postgres: save evidence for finding %q: %w", finding.ID, err)
			}
		}
		if _, err := exec.ExecContext(ctx, `DELETE FROM recommendations WHERE finding_id = $1`, finding.ID); err != nil {
			return fmt.Errorf("postgres: replace recommendations for finding %q: %w", finding.ID, err)
		}
		for _, recommendation := range finding.Recommendations {
			if _, err := exec.ExecContext(ctx, `
INSERT INTO recommendations (finding_id, title, action, rationale, priority)
VALUES ($1, $2, $3, $4, $5)`,
				finding.ID, recommendation.Title, recommendation.Action, recommendation.Rationale, string(recommendation.Priority)); err != nil {
				return fmt.Errorf("postgres: save recommendation for finding %q: %w", finding.ID, err)
			}
		}
	}
	return nil
}

// saveAssessment upserts one assessment row for a scan.
func saveAssessment(ctx context.Context, exec dbExecutor, scanID string, value assessment.Assessment) error {
	severityCounts, err := marshalJSON(stringKeySeverityCounts(value.SeverityCounts))
	if err != nil {
		return fmt.Errorf("postgres: save assessment: marshal severity counts: %w", err)
	}
	categoryScores, err := marshalJSON(stringKeyCategoryScores(value.CategoryScores))
	if err != nil {
		return fmt.Errorf("postgres: save assessment: marshal category scores: %w", err)
	}
	categoryBreakdown, err := marshalJSON(stringKeyCategoryBreakdown(value.CategoryBreakdown))
	if err != nil {
		return fmt.Errorf("postgres: save assessment: marshal category breakdown: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `
INSERT INTO assessments (
	scan_id, overall_score, label, finding_count, severity_counts, category_scores, category_breakdown
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (scan_id) DO UPDATE SET
	overall_score = EXCLUDED.overall_score,
	label = EXCLUDED.label,
	finding_count = EXCLUDED.finding_count,
	severity_counts = EXCLUDED.severity_counts,
	category_scores = EXCLUDED.category_scores,
	category_breakdown = EXCLUDED.category_breakdown`,
		scanID,
		value.OverallScore,
		string(value.Label),
		value.FindingCount,
		severityCounts,
		categoryScores,
		categoryBreakdown,
	); err != nil {
		return databaseError("save assessment", err)
	}
	return nil
}

// saveMetricSnapshots replaces derived metric snapshots for a scan.
func saveMetricSnapshots(ctx context.Context, exec dbExecutor, scanID string, value assessment.Assessment) error {
	if _, err := exec.ExecContext(ctx, `DELETE FROM metric_snapshots WHERE scan_id = $1`, scanID); err != nil {
		return fmt.Errorf("postgres: replace metric snapshots: %w", err)
	}
	metrics := map[string]int{
		"assessment.overall_score": value.OverallScore,
		"findings.total":           value.FindingCount,
	}
	for key, metricValue := range metrics {
		if _, err := exec.ExecContext(ctx, `
INSERT INTO metric_snapshots (scan_id, metric_key, value)
VALUES ($1, $2, $3)`, scanID, key, metricValue); err != nil {
			return fmt.Errorf("postgres: save metric %q: %w", key, err)
		}
	}
	return nil
}

// saveReportMetadata replaces report metadata rows for supported report formats.
func saveReportMetadata(ctx context.Context, exec dbExecutor, scanID string) error {
	if _, err := exec.ExecContext(ctx, `DELETE FROM reports WHERE scan_id = $1`, scanID); err != nil {
		return fmt.Errorf("postgres: replace report metadata: %w", err)
	}
	reportRows := []struct {
		format      string
		contentType string
	}{
		{"markdown", "text/markdown; charset=utf-8"},
		{"json", "application/json"},
	}
	for _, row := range reportRows {
		metadata, err := marshalJSON(map[string]string{"schema_version": "1"})
		if err != nil {
			return fmt.Errorf("postgres: save report metadata: marshal metadata: %w", err)
		}
		if _, err := exec.ExecContext(ctx, `
INSERT INTO reports (scan_id, format, content_type, metadata)
VALUES ($1, $2, $3, $4)`, scanID, row.format, row.contentType, metadata); err != nil {
			return fmt.Errorf("postgres: save report metadata %q: %w", row.format, err)
		}
	}
	return nil
}

// collectFindings flattens analyzer result findings.
func collectFindings(results []analyzer.AnalyzerResult) []findings.Finding {
	var collected []findings.Finding
	for _, result := range results {
		collected = append(collected, result.Findings...)
	}
	return collected
}

// marshalJSON serializes arbitrary metadata for JSONB columns.
func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(value)
}

// stringKeySeverityCounts converts severity-keyed maps to JSON-friendly keys.
func stringKeySeverityCounts(values map[rules.Severity]int) map[string]int {
	out := make(map[string]int, len(values))
	for key, value := range values {
		out[string(key)] = value
	}
	return out
}

// stringKeyCategoryScores converts category-keyed maps to JSON-friendly keys.
func stringKeyCategoryScores(values map[rules.Category]int) map[string]int {
	out := make(map[string]int, len(values))
	for key, value := range values {
		out[string(key)] = value
	}
	return out
}

// stringKeyCategoryBreakdown converts category breakdown maps to JSON-friendly keys.
func stringKeyCategoryBreakdown(values map[rules.Category]assessment.CategoryBreakdown) map[string]assessment.CategoryBreakdown {
	out := make(map[string]assessment.CategoryBreakdown, len(values))
	for key, value := range values {
		out[string(key)] = value
	}
	return out
}

// GetActiveAssessmentPolicy fetches the active assessment policy for an org.
func (s *Store) GetActiveAssessmentPolicy(ctx context.Context, orgID string) (assessment.OrgPolicy, error) {
	// For MVP, we will query the policies table for a policy named "assessment_policy".
	// If not found, return empty policy.
	query := `
		SELECT rules FROM policies
		WHERE organization_id = $1 AND name = 'assessment_policy' AND status = 'active'
		ORDER BY version DESC, updated_at DESC LIMIT 1
	`
	var rulesData []byte
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&rulesData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return assessment.OrgPolicy{}, nil
		}
		return assessment.OrgPolicy{}, err
	}

	var policy assessment.OrgPolicy
	if err := json.Unmarshal(rulesData, &policy); err != nil {
		return assessment.OrgPolicy{}, err
	}
	return policy, nil
}

// databaseError maps database and context failures to stable RepoCompass error codes.
func databaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	code := rcerr.CodeDatabaseQueryFailed
	message := operation + " failed"
	if errors.Is(err, sql.ErrNoRows) {
		code = rcerr.CodeDatabaseNotFound
		message = operation + " not found"
	} else if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		code = rcerr.CodeDatabaseTimeout
		message = operation + " timed out"
	} else {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503", "23505", "23514", "23502":
				code = rcerr.CodeDatabaseConstraintFailed
			case "08000", "08001", "08003", "08006":
				code = rcerr.CodeDatabaseUnavailable
			}
		}
	}
	return rcerr.New(code, message, err)
}

// validationError returns a stable database query error before invalid data reaches Postgres.
func validationError(message string) error {
	return rcerr.New(rcerr.CodeDatabaseQueryFailed, message, nil)
}

// validateRunResult verifies required persisted aggregate fields before writing.
func validateRunResult(result scan.RunResult) error {
	if err := validateRepository(result.Repository); err != nil {
		return err
	}
	if err := validateSnapshot(result.Snapshot); err != nil {
		return err
	}
	if err := validateScan(result.Scan); err != nil {
		return err
	}
	if result.Snapshot.RepositoryID != result.Repository.ID {
		return validationError("snapshot repository id must match repository id")
	}
	if result.Scan.SnapshotID != result.Snapshot.ID {
		return validationError("scan snapshot id must match snapshot id")
	}
	for _, analyzerResult := range result.AnalyzerResults {
		if analyzerResult.AnalyzerID == "" {
			return validationError("analyzer result id is required")
		}
		if analyzerResult.Name == "" {
			return validationError("analyzer result name is required")
		}
		for _, finding := range analyzerResult.Findings {
			if err := validateFinding(finding); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateRepository verifies required repository row fields.
func validateRepository(repo repository.Repository) error {
	if repo.ID == "" {
		return validationError("repository id is required")
	}
	if repo.Name == "" {
		return validationError("repository name is required")
	}
	if repo.FullName == "" {
		return validationError("repository full name is required")
	}
	if repo.Provider == "" {
		return validationError("repository provider is required")
	}
	if repo.Status == "" {
		return validationError("repository status is required")
	}
	return nil
}

// validateSnapshot verifies required snapshot row fields.
func validateSnapshot(snap snapshot.RepositorySnapshot) error {
	if snap.ID == "" {
		return validationError("snapshot id is required")
	}
	if snap.RepositoryID == "" {
		return validationError("snapshot repository id is required")
	}
	if snap.SourceType == "" {
		return validationError("snapshot source type is required")
	}
	if snap.CapturedAt.IsZero() {
		return validationError("snapshot captured_at is required")
	}
	return nil
}

// validateScan verifies required scan row fields.
func validateScan(sc scan.Scan) error {
	if sc.ID == "" {
		return validationError("scan id is required")
	}
	if sc.SnapshotID == "" {
		return validationError("scan snapshot id is required")
	}
	if sc.Status == "" {
		return validationError("scan status is required")
	}
	return nil
}

// validateFinding verifies required finding row fields.
func validateFinding(finding findings.Finding) error {
	if finding.ID == "" {
		return validationError("finding id is required")
	}
	if finding.RuleID == "" {
		return validationError("finding rule id is required")
	}
	if finding.AnalyzerID == "" {
		return validationError("finding analyzer id is required")
	}
	if finding.Severity == "" {
		return validationError("finding severity is required")
	}
	if finding.Title == "" {
		return validationError("finding title is required")
	}
	if finding.Message == "" {
		return validationError("finding message is required")
	}
	if finding.Category == "" {
		return validationError("finding category is required")
	}
	return nil
}
