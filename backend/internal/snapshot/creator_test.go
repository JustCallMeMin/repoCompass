package snapshot

import (
	"context"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

func TestCreatorCreatesSnapshotFromRepositoryMetadata(t *testing.T) {
	capturedAt := time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
	creator := Creator{
		now: func() time.Time {
			return capturedAt
		},
	}

	snapshot, err := creator.Create(context.Background(), CreateRequest{
		Repository: repository.Repository{
			ID:            "local_abc123",
			DefaultBranch: "main",
		},
		SourceType: SourceTypeLocal,
		SnapshotMetadata: map[string]string{
			"default_branch": "trunk",
			"commit_sha":     "abc123",
			"tree_reference": "tree-abc123",
			"local_path":     "/tmp/basic-go-repo",
		},
	})
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	if snapshot.ID == "" {
		t.Fatal("expected snapshot ID to be set")
	}
	if snapshot.RepositoryID != "local_abc123" {
		t.Fatalf("unexpected repository ID: got %q", snapshot.RepositoryID)
	}
	if snapshot.SourceType != SourceTypeLocal {
		t.Fatalf("unexpected source type: got %q", snapshot.SourceType)
	}
	if snapshot.BranchName != "trunk" {
		t.Fatalf("unexpected branch name: got %q", snapshot.BranchName)
	}
	if snapshot.CommitSHA != "abc123" {
		t.Fatalf("unexpected commit SHA: got %q", snapshot.CommitSHA)
	}
	if snapshot.TreeReference != "tree-abc123" {
		t.Fatalf("unexpected tree reference: got %q", snapshot.TreeReference)
	}
	if !snapshot.CapturedAt.Equal(capturedAt) {
		t.Fatalf("unexpected captured time: got %s want %s", snapshot.CapturedAt, capturedAt)
	}
	if snapshot.SnapshotMetadata["local_path"] != "/tmp/basic-go-repo" {
		t.Fatalf("unexpected metadata local_path: got %q", snapshot.SnapshotMetadata["local_path"])
	}
}

func TestCreatorFallsBackToRepositoryDefaultBranch(t *testing.T) {
	creator := Creator{
		now: func() time.Time {
			return time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
		},
	}

	snapshot, err := creator.Create(context.Background(), CreateRequest{
		Repository: repository.Repository{
			ID:            "local_abc123",
			DefaultBranch: "main",
		},
		SourceType: SourceTypeLocal,
	})
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}
	if snapshot.BranchName != "main" {
		t.Fatalf("unexpected branch fallback: got %q", snapshot.BranchName)
	}
}

func TestCreatorCopiesSnapshotMetadata(t *testing.T) {
	metadata := map[string]string{
		"local_path": "/tmp/basic-go-repo",
	}
	creator := Creator{
		now: func() time.Time {
			return time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
		},
	}

	snapshot, err := creator.Create(context.Background(), CreateRequest{
		Repository:       repository.Repository{ID: "local_abc123"},
		SourceType:       SourceTypeLocal,
		SnapshotMetadata: metadata,
	})
	if err != nil {
		t.Fatalf("expected snapshot creation to succeed: %v", err)
	}

	metadata["local_path"] = "/tmp/mutated"
	if snapshot.SnapshotMetadata["local_path"] != "/tmp/basic-go-repo" {
		t.Fatalf("expected snapshot metadata copy, got %q", snapshot.SnapshotMetadata["local_path"])
	}
}

func TestCreatorRejectsMissingRepositoryID(t *testing.T) {
	creator := NewCreator()

	_, err := creator.Create(context.Background(), CreateRequest{
		SourceType: SourceTypeLocal,
	})
	if err == nil {
		t.Fatal("expected missing repository ID to fail")
	}
}

func TestCreatorRejectsEmptySourceType(t *testing.T) {
	creator := NewCreator()

	_, err := creator.Create(context.Background(), CreateRequest{
		Repository: repository.Repository{ID: "local_abc123"},
	})
	if err == nil {
		t.Fatal("expected empty source type to fail")
	}
}
