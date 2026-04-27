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

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
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

	for _, tableName := range []string{"repositories", "repository_snapshots", "scans"} {
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
