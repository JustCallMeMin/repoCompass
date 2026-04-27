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
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
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
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}

// New creates a Store backed by the provided *sql.DB.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// SaveRepository upserts a repository row identified by its ID.
// If a row with the same ID already exists, all mutable fields are updated.
func (s *Store) SaveRepository(ctx context.Context, repo repository.Repository) error {
	const q = `
INSERT INTO repositories (
	id, name, owner_name, full_name, url, provider,
	default_branch, primary_ecosystem, is_monorepo, status,
	organization_id, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6,
	$7, $8, $9, $10,
	$11, NOW()
)
ON CONFLICT (id) DO UPDATE SET
	name              = EXCLUDED.name,
	owner_name        = EXCLUDED.owner_name,
	full_name         = EXCLUDED.full_name,
	url               = EXCLUDED.url,
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
		string(repo.Provider),
		repo.DefaultBranch,
		repo.PrimaryEcosystem,
		repo.IsMonorepo,
		string(repo.Status),
		repo.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("postgres: SaveRepository: %w", err)
	}
	return nil
}

// SaveSnapshot inserts a new repository_snapshot row.
func (s *Store) SaveSnapshot(ctx context.Context, snap snapshot.RepositorySnapshot) error {
	metadata, err := marshalMetadata(snap.SnapshotMetadata)
	if err != nil {
		return fmt.Errorf("postgres: SaveSnapshot: marshal metadata: %w", err)
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
		return fmt.Errorf("postgres: SaveSnapshot: %w", err)
	}
	return nil
}

// SaveScan inserts a new scan row with its initial state.
func (s *Store) SaveScan(ctx context.Context, sc scan.Scan) error {
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
		return fmt.Errorf("postgres: SaveScan: %w", err)
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
		return fmt.Errorf("postgres: UpdateScan: %w", err)
	}
	return nil
}

// marshalMetadata serialises a metadata map to a JSON byte slice suitable for
// storing in a JSONB column. A nil map marshals to the JSON object "{}".
func marshalMetadata(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
