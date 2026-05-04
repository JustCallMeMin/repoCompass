package scan

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Store defines the persistence operations required by the scan lifecycle.
//
// Implementations are provided by internal/storage/noop and internal/storage/postgres.
// Pass noop.New() when persistence is disabled (default CLI mode).
type Store interface {
	// SaveRepository upserts a repository record. Safe to call on repeat scans.
	SaveRepository(ctx context.Context, repo repository.Repository) error

	// SaveSnapshot inserts a new snapshot record.
	SaveSnapshot(ctx context.Context, snap snapshot.RepositorySnapshot) error

	// SaveScan inserts a new scan record with its initial state.
	SaveScan(ctx context.Context, s Scan) error

	// UpdateScan updates mutable scan fields (status, end_time, error_details).
	UpdateScan(ctx context.Context, s Scan) error

	// SaveRunResult persists the completed scan result as one logical unit.
	SaveRunResult(ctx context.Context, result RunResult) error
}
