package scan

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// LocalScanRunner orchestrates the local scan execution lifecycle.
type LocalScanRunner struct {
	provider repository.LocalRepositoryProvider
	creator  snapshot.Creator
}

// NewLocalScanRunner creates a new LocalScanRunner.
func NewLocalScanRunner(provider repository.LocalRepositoryProvider, creator snapshot.Creator) *LocalScanRunner {
	return &LocalScanRunner{
		provider: provider,
		creator:  creator,
	}
}

// Run executes a scan locally by resolving the repository, creating a snapshot, and returning a summary.
func (r *LocalScanRunner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	s := Scan{
		ID:     generateScanID(),
		Status: StatusCreated,
	}

	// Transition to running
	if err := s.TransitionTo(StatusRunning); err != nil {
		return RunResult{Scan: s}, fmt.Errorf("failed to start scan: %w", err)
	}

	now := time.Now()
	s.StartTime = &now

	// 1. Resolve Repository
	resolution, err := r.provider.Resolve(ctx, request.Source)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed) // Ignore transition error as actual error takes precedence
		return RunResult{Scan: s}, fmt.Errorf("failed to resolve repository: %w", err)
	}

	// 2. Create Snapshot
	snap, err := r.creator.Create(ctx, snapshot.CreateRequest{
		Repository:       resolution.Repository,
		SourceType:       snapshot.SourceTypeLocal,
		SnapshotMetadata: resolution.SnapshotMetadata,
	})
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		return RunResult{Scan: s}, fmt.Errorf("failed to create snapshot: %w", err)
	}

	s.SnapshotID = snap.ID

	// 3. Complete Scan (no-op analysis for Milestone 1)
	if err := s.TransitionTo(StatusCompleted); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		return RunResult{Scan: s}, fmt.Errorf("failed to complete scan: %w", err)
	}

	endTime := time.Now()
	s.EndTime = &endTime

	return RunResult{
		Scan: s,
		Summary: Summary{
			AnalyzersProcessed: 0,
		},
	}, nil
}

func generateScanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "scan_" + hex.EncodeToString(b)
}
