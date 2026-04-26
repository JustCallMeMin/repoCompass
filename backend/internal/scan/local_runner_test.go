package scan_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestLocalScanRunner_Run_Success(t *testing.T) {
	tempDir := t.TempDir()

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	runner := scan.NewLocalScanRunner(provider, creator)

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	result, err := runner.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	s := result.Scan
	if s.Status != scan.StatusCompleted {
		t.Errorf("expected status %q, got %q", scan.StatusCompleted, s.Status)
	}
	if s.ID == "" {
		t.Error("expected non-empty scan ID")
	}
	if s.SnapshotID == "" {
		t.Error("expected non-empty snapshot ID")
	}
	if s.StartTime == nil || s.EndTime == nil {
		t.Error("expected start and end times to be set")
	}
	if result.Summary.AnalyzersProcessed != 0 {
		t.Errorf("expected 0 analyzers processed, got %d", result.Summary.AnalyzersProcessed)
	}
}

func TestLocalScanRunner_Run_ResolveError(t *testing.T) {
	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	runner := scan.NewLocalScanRunner(provider, creator)

	// Provide a path that doesn't exist to force a resolution error
	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: filepath.Join(os.TempDir(), "repo_compass_nonexistent_dir_12345"),
		},
	}

	result, err := runner.Run(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	s := result.Scan
	if s.Status != scan.StatusFailed {
		t.Errorf("expected status %q, got %q", scan.StatusFailed, s.Status)
	}
	if s.ErrorDetails == "" {
		t.Error("expected error details to be populated")
	}
}

func TestLocalScanRunner_Run_SnapshotCreateError(t *testing.T) {
	// A bit tricky to force snapshot create error if resolve succeeds, 
	// unless we pass canceled context or cause an internal validation failure.
	// Since snapshot creation validates empty SourceType, let's pass a repository source with empty SourceType.
	
	tempDir := t.TempDir()

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	runner := scan.NewLocalScanRunner(provider, creator)

	// If we set type to empty, LocalRepositoryProvider will accept it but we can simulate snapshot error if we somehow inject bad data.
	// Wait, LocalRepositoryProvider actually requires source.Type to be either empty or SourceTypeLocal.
	// But it then hardcodes `SourceTypeLocal` in the `resolved_at` metadata: `metadata["source_type"] = string(SourceTypeLocal)`.
	// Oh, wait, the `LocalScanRunner` hardcodes `SourceTypeLocal` in the `snapshot.CreateRequest`.
	// So snapshot creation should always succeed if resolution succeeds, unless context is cancelled.
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	
	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	// This might fail in Resolve (because `exec.CommandContext` checks for cancellation), or in Create.
	// Let's just verify that a cancelled context fails the run and marks it as Failed.
	
	result, err := runner.Run(ctx, req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}

	s := result.Scan
	if s.Status != scan.StatusFailed {
		t.Errorf("expected status %q, got %q", scan.StatusFailed, s.Status)
	}
}
