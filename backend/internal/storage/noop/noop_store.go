// Package noop provides a no-op implementation of scan.Store.
// It is used when persistence is disabled (i.e. --persist is not passed).
// Every method is a no-op and always returns nil.
package noop

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Store is a no-op implementation of scan.Store.
// It silently discards all persistence calls and never returns an error.
type Store struct{}

// New returns a new no-op Store.
func New() *Store {
	return &Store{}
}

// SaveRepository is a no-op.
func (s *Store) SaveRepository(_ context.Context, _ repository.Repository) error {
	return nil
}

// SaveSnapshot is a no-op.
func (s *Store) SaveSnapshot(_ context.Context, _ snapshot.RepositorySnapshot) error {
	return nil
}

// SaveScan is a no-op.
func (s *Store) SaveScan(_ context.Context, _ scan.Scan) error {
	return nil
}

// UpdateScan is a no-op.
func (s *Store) UpdateScan(_ context.Context, _ scan.Scan) error {
	return nil
}
