package scan

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// LocalScanRunner orchestrates the local scan execution lifecycle.
type LocalScanRunner struct {
	provider repository.LocalRepositoryProvider
	creator  snapshot.Creator
	resolver config.Resolver
	logger   *slog.Logger
	store    Store
}

// NewLocalScanRunner creates a new LocalScanRunner.
// logger is required; pass slog.Default() if no specific logger is needed.
// store is required; pass noop.New() if persistence is disabled.
func NewLocalScanRunner(
	provider repository.LocalRepositoryProvider,
	creator snapshot.Creator,
	resolver config.Resolver,
	logger *slog.Logger,
	store Store,
) *LocalScanRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &LocalScanRunner{
		provider: provider,
		creator:  creator,
		resolver: resolver,
		logger:   logger,
		store:    store,
	}
}

// Run executes a scan locally by resolving the repository, creating a snapshot, and returning a summary.
func (r *LocalScanRunner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	s := Scan{
		ID:     generateScanID(),
		Status: StatusCreated,
	}

	// Bind scan_id to every log call in this run.
	l := r.logger.With(
		slog.String("scan_id", s.ID),
		slog.String("repo_path", request.Source.Path),
	)

	// Transition to running
	if err := s.TransitionTo(StatusRunning); err != nil {
		return RunResult{Scan: s}, rcerr.New(rcerr.CodeInternalError, "failed to start scan", err)
	}

	start := time.Now()
	s.StartTime = &start

	l.InfoContext(ctx, "scan started",
		slog.String("operation", "scan_start"),
	)

	// 0. Resolve Configuration
	effConfig, err := r.resolver.ResolveConfig(ctx, request.Source.Path, request.ConfigOverrides)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := wrapIfNeeded(err, rcerr.CodeConfigResolveFailed, "failed to resolve configuration")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	l.InfoContext(ctx, "config resolved",
		slog.String("operation", "config_resolve"),
		slog.Int64("max_file_size_bytes", effConfig.MaxFileSizeBytes),
	)

	// 1. Resolve Repository
	resolution, err := r.provider.Resolve(ctx, request.Source)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := wrapIfNeeded(err, rcerr.CodeRepoResolveFailed, "failed to resolve repository")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	l.InfoContext(ctx, "repository resolved",
		slog.String("operation", "repo_resolve"),
		slog.String("repository_id", resolution.Repository.ID),
		slog.String("repository_name", resolution.Repository.Name),
	)

	// Persist repository (upsert — safe to call on repeat scans of same repo).
	if err := r.store.SaveRepository(ctx, resolution.Repository); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to persist repository", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
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
		wrapped := wrapIfNeeded(err, rcerr.CodeSnapshotCreateFailed, "failed to create snapshot")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	s.SnapshotID = snap.ID

	l.InfoContext(ctx, "snapshot created",
		slog.String("operation", "snapshot_create"),
		slog.String("snapshot_id", snap.ID),
		slog.String("repository_id", resolution.Repository.ID),
	)

	// Persist snapshot.
	if err := r.store.SaveSnapshot(ctx, snap); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to persist snapshot", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	// Persist scan (initial running state).
	if err := r.store.SaveScan(ctx, s); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to persist scan", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	// 3. Complete Scan (no-op analysis for Milestone 1)
	if err := s.TransitionTo(StatusCompleted); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to complete scan", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	endTime := time.Now()
	s.EndTime = &endTime

	// Persist final scan state.
	if err := r.store.UpdateScan(ctx, s); err != nil {
		l.ErrorContext(ctx, "failed to update scan in store",
			slog.String("operation", "scan_completed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_msg", err.Error()),
		)
		// UpdateScan failure after a successful scan is surfaced as a scan error.
		return RunResult{Scan: s}, rcerr.New(rcerr.CodeScanExecutionFailed, "failed to update scan record", err)
	}

	l.InfoContext(ctx, "scan completed",
		slog.String("operation", "scan_completed"),
		slog.String("repository_id", resolution.Repository.ID),
		slog.String("snapshot_id", snap.ID),
		slog.String("status", string(StatusCompleted)),
		slog.Int64("duration_ms", endTime.Sub(start).Milliseconds()),
	)

	return RunResult{
		Scan: s,
		Summary: Summary{
			AnalyzersProcessed: 0,
		},
		EffectiveConfig: effConfig,
	}, nil
}

func generateScanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "scan_" + hex.EncodeToString(b)
}

// wrapIfNeeded wraps err with a new rcerr.Error only if err is not already an *rcerr.Error.
// This prevents double-wrapping when sub-packages already return structured errors.
func wrapIfNeeded(err error, code rcerr.ErrorCode, message string) error {
	var rcErr *rcerr.Error
	if errors.As(err, &rcErr) {
		return err
	}
	return rcerr.New(code, message, err)
}
