package snapshot

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrSnapshotNotFound is returned when a snapshot ID is not present in a store.
var ErrSnapshotNotFound = errors.New("snapshot not found")

// SnapshotStore persists and retrieves repository snapshots.
type SnapshotStore interface {
	Save(ctx context.Context, snapshot RepositorySnapshot) error
	Get(ctx context.Context, id string) (RepositorySnapshot, error)
}

// InMemorySnapshotStore stores snapshots in process memory for early development.
type InMemorySnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string]RepositorySnapshot
}

// NewInMemorySnapshotStore creates an empty in-memory snapshot store.
func NewInMemorySnapshotStore() *InMemorySnapshotStore {
	return &InMemorySnapshotStore{
		snapshots: make(map[string]RepositorySnapshot),
	}
}

// Save stores or replaces a snapshot by ID.
func (s *InMemorySnapshotStore) Save(ctx context.Context, snapshot RepositorySnapshot) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if snapshot.ID == "" {
		return fmt.Errorf("snapshot ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snapshot.ID] = cloneSnapshot(snapshot)
	return nil
}

// Get retrieves a snapshot by ID.
func (s *InMemorySnapshotStore) Get(ctx context.Context, id string) (RepositorySnapshot, error) {
	select {
	case <-ctx.Done():
		return RepositorySnapshot{}, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[id]
	if !ok {
		return RepositorySnapshot{}, ErrSnapshotNotFound
	}

	return cloneSnapshot(snapshot), nil
}

func cloneSnapshot(snapshot RepositorySnapshot) RepositorySnapshot {
	snapshot.SnapshotMetadata = copyMetadata(snapshot.SnapshotMetadata)
	return snapshot
}
