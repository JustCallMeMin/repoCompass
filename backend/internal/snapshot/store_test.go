package snapshot

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInMemorySnapshotStoreSavesAndGetsSnapshot(t *testing.T) {
	store := NewInMemorySnapshotStore()
	expected := RepositorySnapshot{
		ID:           "snap_123",
		RepositoryID: "repo_123",
		SourceType:   SourceTypeLocal,
		CapturedAt:   time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC),
		SnapshotMetadata: map[string]string{
			"local_path": "/tmp/basic-go-repo",
		},
	}

	if err := store.Save(context.Background(), expected); err != nil {
		t.Fatalf("expected save to succeed: %v", err)
	}

	got, err := store.Get(context.Background(), expected.ID)
	if err != nil {
		t.Fatalf("expected get to succeed: %v", err)
	}

	if got.ID != expected.ID {
		t.Fatalf("unexpected snapshot ID: got %q want %q", got.ID, expected.ID)
	}
	if got.RepositoryID != expected.RepositoryID {
		t.Fatalf("unexpected repository ID: got %q want %q", got.RepositoryID, expected.RepositoryID)
	}
	if got.SnapshotMetadata["local_path"] != expected.SnapshotMetadata["local_path"] {
		t.Fatalf("unexpected metadata: got %q", got.SnapshotMetadata["local_path"])
	}
}

func TestInMemorySnapshotStoreReturnsCopy(t *testing.T) {
	store := NewInMemorySnapshotStore()
	snapshot := RepositorySnapshot{
		ID: "snap_123",
		SnapshotMetadata: map[string]string{
			"local_path": "/tmp/basic-go-repo",
		},
	}

	if err := store.Save(context.Background(), snapshot); err != nil {
		t.Fatalf("expected save to succeed: %v", err)
	}

	snapshot.SnapshotMetadata["local_path"] = "/tmp/mutated-before-get"
	got, err := store.Get(context.Background(), snapshot.ID)
	if err != nil {
		t.Fatalf("expected get to succeed: %v", err)
	}
	got.SnapshotMetadata["local_path"] = "/tmp/mutated-after-get"

	gotAgain, err := store.Get(context.Background(), snapshot.ID)
	if err != nil {
		t.Fatalf("expected second get to succeed: %v", err)
	}
	if gotAgain.SnapshotMetadata["local_path"] != "/tmp/basic-go-repo" {
		t.Fatalf("expected stored snapshot metadata copy, got %q", gotAgain.SnapshotMetadata["local_path"])
	}
}

func TestInMemorySnapshotStoreRejectsEmptySnapshotID(t *testing.T) {
	store := NewInMemorySnapshotStore()

	err := store.Save(context.Background(), RepositorySnapshot{})
	if err == nil {
		t.Fatal("expected empty snapshot ID to fail")
	}
}

func TestInMemorySnapshotStoreReturnsNotFound(t *testing.T) {
	store := NewInMemorySnapshotStore()

	_, err := store.Get(context.Background(), "missing")
	if !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
