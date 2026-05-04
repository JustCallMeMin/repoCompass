// Package postgres_test provides integration tests for the postgres.Store.
// Tests are skipped automatically when DATABASE_URL is not set, so they do not
// require a database in the standard CI/unit-test flow.
//
// To run manually:
//
//	export DATABASE_URL="postgres://postgres:postgres@localhost:5432/repocompass?sslmode=disable"
//	go test ./internal/storage/postgres/...
package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
)

// openDB opens a database connection for integration tests.
// The test is skipped when DATABASE_URL is unset.
func openDB(t *testing.T) *pgstore.Store {
	return pgstore.New(openDBConn(t))
}

func openDBConn(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping postgres integration test")
	}
	db, err := pgstore.Open(dsn)
	if err != nil {
		t.Fatalf("postgres: Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestPostgresStore_CoreTablesExist(t *testing.T) {
	db := openDBConn(t)
	ctx := context.Background()

	for _, tableName := range []string{
		"repositories",
		"repository_snapshots",
		"scans",
		"rule_sets",
		"rules",
		"rule_set_rules",
		"analyzer_results",
		"findings",
		"finding_evidences",
		"recommendations",
		"assessments",
		"metric_snapshots",
		"reports",
	} {
		var exists bool
		err := db.QueryRowContext(ctx, `
SELECT EXISTS (
	SELECT 1
	FROM information_schema.tables
	WHERE table_schema = 'public'
	  AND table_name = $1
)`, tableName).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", tableName, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist after migrations", tableName)
		}
	}
}

func TestPostgresStore_SaveRepository_Upsert(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()
	repoID := testID("test-repo-upsert")

	repo := repository.Repository{
		ID:        repoID,
		Name:      "my-repo",
		OwnerName: "alice",
		FullName:  "alice/my-repo",
		Provider:  repository.ProviderLocal,
		Status:    repository.StatusActive,
	}

	// First save — insert.
	if err := store.SaveRepository(ctx, repo); err != nil {
		t.Fatalf("first SaveRepository: %v", err)
	}

	// Second save — upsert must not fail.
	repo.Name = "my-repo-renamed"
	if err := store.SaveRepository(ctx, repo); err != nil {
		t.Fatalf("second SaveRepository (upsert): %v", err)
	}
}

func TestPostgresStore_SaveSnapshot(t *testing.T) {
	db := openDBConn(t)
	store := pgstore.New(db)
	ctx := context.Background()
	repoID := testID("test-repo-snap")
	snapshotID := testID("test-snap")

	// Seed required parent repository.
	repo := repository.Repository{
		ID:       repoID,
		Name:     "snap-repo",
		FullName: "alice/snap-repo",
		Provider: repository.ProviderLocal,
		Status:   repository.StatusActive,
	}
	if err := store.SaveRepository(ctx, repo); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	snap := snapshot.RepositorySnapshot{
		ID:           snapshotID,
		RepositoryID: repo.ID,
		SourceType:   snapshot.SourceTypeLocal,
		CapturedAt:   time.Now().UTC().Truncate(time.Second),
		SnapshotMetadata: map[string]string{
			"branch": "main",
		},
	}
	if err := store.SaveSnapshot(ctx, snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	var branch string
	err := db.QueryRowContext(ctx, `
SELECT snapshot_metadata->>'branch'
FROM repository_snapshots
WHERE id = $1`, snapshotID).Scan(&branch)
	if err != nil {
		t.Fatalf("query snapshot metadata: %v", err)
	}
	if branch != "main" {
		t.Fatalf("unexpected JSONB metadata branch: got %q want %q", branch, "main")
	}
}

func TestPostgresStore_SaveAndUpdateScan(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()
	repoID := testID("test-repo-scan")
	snapshotID := testID("test-snap-scan")
	scanID := testID("test-scan")

	// Seed required parent repo + snapshot.
	repo := repository.Repository{
		ID:       repoID,
		Name:     "scan-repo",
		FullName: "alice/scan-repo",
		Provider: repository.ProviderLocal,
		Status:   repository.StatusActive,
	}
	if err := store.SaveRepository(ctx, repo); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	snap := snapshot.RepositorySnapshot{
		ID:           snapshotID,
		RepositoryID: repo.ID,
		SourceType:   snapshot.SourceTypeLocal,
		CapturedAt:   now,
	}
	if err := store.SaveSnapshot(ctx, snap); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	s := scan.Scan{
		ID:         scanID,
		SnapshotID: snap.ID,
		Status:     scan.StatusRunning,
		StartTime:  &now,
	}
	if err := store.SaveScan(ctx, s); err != nil {
		t.Fatalf("SaveScan: %v", err)
	}

	endTime := now.Add(2 * time.Second)
	s.Status = scan.StatusCompleted
	s.EndTime = &endTime

	if err := store.UpdateScan(ctx, s); err != nil {
		t.Fatalf("UpdateScan: %v", err)
	}
}

func TestPostgresStore_ForeignKeyFailures(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()

	snap := snapshot.RepositorySnapshot{
		ID:           testID("test-snap-missing-parent"),
		RepositoryID: testID("missing-repo"),
		SourceType:   snapshot.SourceTypeLocal,
		CapturedAt:   time.Now().UTC().Truncate(time.Second),
	}
	if err := store.SaveSnapshot(ctx, snap); err == nil {
		t.Fatal("expected SaveSnapshot to fail when repository_id does not exist")
	}

	now := time.Now().UTC().Truncate(time.Second)
	s := scan.Scan{
		ID:         testID("test-scan-missing-parent"),
		SnapshotID: testID("missing-snapshot"),
		Status:     scan.StatusRunning,
		StartTime:  &now,
	}
	if err := store.SaveScan(ctx, s); err == nil {
		t.Fatal("expected SaveScan to fail when snapshot_id does not exist")
	}
}

func TestPostgresStore_SaveRunResultPersistsAnalysisRows(t *testing.T) {
	db := openDBConn(t)
	store := pgstore.New(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	repoID := testID("test-repo-result")
	scanID := testID("test-scan-result")
	snapshotID := testID("test-snap-result")
	findingID := testID("test-finding-result")
	finding := findings.NewFinding(
		findingID,
		"readme.exists",
		"readme",
		rules.SeverityHigh,
		rules.CategoryDocumentation,
		"README missing",
		"Repository root has no README.",
		[]findings.Evidence{
			findings.NewEvidence(findings.EvidenceTypeFileMissing, "README.md missing.", "README.md", ""),
		},
	)
	finding.Recommendations = []findings.Recommendation{
		findings.NewRecommendation(findingID, "Add README", "Create README.md.", "Contributors need orientation.", findings.RecommendationPriorityHigh),
	}
	result := scan.RunResult{
		Repository: repository.Repository{
			ID:       repoID,
			Name:     "result-repo",
			FullName: "alice/result-repo",
			Provider: repository.ProviderLocal,
			Status:   repository.StatusActive,
		},
		Snapshot: snapshot.RepositorySnapshot{
			ID:           snapshotID,
			RepositoryID: repoID,
			SourceType:   snapshot.SourceTypeLocal,
			CapturedAt:   now,
		},
		Scan: scan.Scan{
			ID:         scanID,
			SnapshotID: snapshotID,
			Status:     scan.StatusCompleted,
			StartTime:  &now,
			EndTime:    ptrTime(now.Add(time.Second)),
		},
		AnalyzerResults: []analyzer.AnalyzerResult{
			{
				AnalyzerID: "readme",
				Name:       "README Analyzer",
				Version:    "0.1.0",
				Status:     analyzer.AnalyzerStatusSuccess,
				Findings:   []findings.Finding{finding},
				Metadata:   map[string]string{"readme_present": "false"},
			},
		},
		Assessment: assessment.Assessment{
			OverallScore:   75,
			Label:          assessment.ScoreLabelGood,
			FindingCount:   1,
			SeverityCounts: map[rules.Severity]int{rules.SeverityHigh: 1},
			CategoryScores: map[rules.Category]int{rules.CategoryDocumentation: 75},
		},
	}

	if err := store.SaveRunResult(ctx, result); err != nil {
		t.Fatalf("SaveRunResult: %v", err)
	}

	counts := map[string]int{}
	for _, query := range []struct {
		name string
		sql  string
	}{
		{"analyzer_results", `SELECT COUNT(*) FROM analyzer_results WHERE scan_id = $1`},
		{"findings", `SELECT COUNT(*) FROM findings WHERE scan_id = $1`},
		{"assessments", `SELECT COUNT(*) FROM assessments WHERE scan_id = $1`},
		{"metrics", `SELECT COUNT(*) FROM metric_snapshots WHERE scan_id = $1`},
		{"reports", `SELECT COUNT(*) FROM reports WHERE scan_id = $1`},
	} {
		var count int
		if err := db.QueryRowContext(ctx, query.sql, scanID).Scan(&count); err != nil {
			t.Fatalf("query %s count: %v", query.name, err)
		}
		counts[query.name] = count
	}
	if counts["analyzer_results"] != 1 || counts["findings"] != 1 || counts["assessments"] != 1 {
		t.Fatalf("unexpected core analysis counts: %v", counts)
	}
	if counts["metrics"] < 2 {
		t.Fatalf("expected metric snapshots, got counts: %v", counts)
	}
	if counts["reports"] != 2 {
		t.Fatalf("expected markdown/json report metadata rows, got counts: %v", counts)
	}

	historyItems, err := store.ListScanHistory(ctx, repoID, 10)
	if err != nil {
		t.Fatalf("ListScanHistory: %v", err)
	}
	if len(historyItems) == 0 || historyItems[0].ScanID != scanID {
		t.Fatalf("expected scan %q in history, got %+v", scanID, historyItems)
	}
	findingItems, err := store.ListFindings(ctx, scanID)
	if err != nil {
		t.Fatalf("ListFindings: %v", err)
	}
	if len(findingItems) != 1 || len(findingItems[0].Evidence) != 1 || len(findingItems[0].Recommendations) != 1 {
		t.Fatalf("expected finding evidence and recommendation, got %+v", findingItems)
	}
	metricPoints, err := store.ListMetricTrend(ctx, repoID, "assessment.overall_score", 10)
	if err != nil {
		t.Fatalf("ListMetricTrend: %v", err)
	}
	if len(metricPoints) == 0 || metricPoints[0].Value != 75 {
		t.Fatalf("expected assessment score metric, got %+v", metricPoints)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
