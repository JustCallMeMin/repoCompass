package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestGetOrganizationInsights(t *testing.T) {
	ctx := context.Background()
	store := openDB(t)

	orgID := "test_org_1"

	// Create repository 1
	repo1 := repository.Repository{
		ID:             "repo_1",
		Name:           "repo1",
		FullName:       "org/repo1",
		Provider:       repository.ProviderLocal,
		Status:         repository.StatusActive,
		OrganizationID: orgID,
	}
	if err := store.SaveRepository(ctx, repo1); err != nil {
		t.Fatalf("SaveRepository failed: %v", err)
	}

	snap1 := snapshot.RepositorySnapshot{
		ID:           "snap_1",
		RepositoryID: "repo_1",
		SourceType:   snapshot.SourceTypeLocal,
		CapturedAt:   time.Now(),
	}
	if err := store.SaveSnapshot(ctx, snap1); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	scan1 := scan.Scan{
		ID:         "scan_1",
		SnapshotID: "snap_1",
		Status:     scan.StatusCompleted,
	}
	if err := store.SaveScan(ctx, scan1); err != nil {
		t.Fatalf("SaveScan failed: %v", err)
	}

	result1 := scan.RunResult{
		Scan:       scan1,
		Repository: repo1,
		Snapshot:   snap1,
		Assessment: assessment.Assessment{OverallScore: 80},
	}
	if err := store.SaveRunResult(ctx, result1); err != nil {
		t.Fatalf("SaveRunResult failed: %v", err)
	}

	// Create repository 2
	repo2 := repository.Repository{
		ID:             "repo_2",
		Name:           "repo2",
		FullName:       "org/repo2",
		Provider:       repository.ProviderLocal,
		Status:         repository.StatusActive,
		OrganizationID: orgID,
	}
	if err := store.SaveRepository(ctx, repo2); err != nil {
		t.Fatalf("SaveRepository failed: %v", err)
	}

	snap2 := snapshot.RepositorySnapshot{
		ID:           "snap_2",
		RepositoryID: "repo_2",
		SourceType:   snapshot.SourceTypeLocal,
		CapturedAt:   time.Now(),
	}
	if err := store.SaveSnapshot(ctx, snap2); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	scan2 := scan.Scan{
		ID:         "scan_2",
		SnapshotID: "snap_2",
		Status:     scan.StatusCompleted,
	}
	if err := store.SaveScan(ctx, scan2); err != nil {
		t.Fatalf("SaveScan failed: %v", err)
	}

	result2 := scan.RunResult{
		Scan:       scan2,
		Repository: repo2,
		Snapshot:   snap2,
		Assessment: assessment.Assessment{OverallScore: 100},
	}
	if err := store.SaveRunResult(ctx, result2); err != nil {
		t.Fatalf("SaveRunResult failed: %v", err)
	}

	// Fetch insights
	insights, err := store.GetOrganizationInsights(ctx, orgID)
	if err != nil {
		t.Fatalf("GetOrganizationInsights failed: %v", err)
	}

	if insights.TotalRepositories != 2 {
		t.Errorf("expected 2 repos, got %d", insights.TotalRepositories)
	}
	if insights.TotalScans != 2 {
		t.Errorf("expected 2 scans, got %d", insights.TotalScans)
	}
	if insights.AverageScore != 90 {
		t.Errorf("expected avg score 90, got %d", insights.AverageScore)
	}
}
