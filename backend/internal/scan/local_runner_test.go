package scan_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
	"github.com/JustCallMeMin/repoCompass/backend/internal/storage/noop"
)

// discardLogger returns a logger that discards all output, suitable for tests
// that do not assert on log output.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// jsonLogger returns a logger that writes JSON lines to buf.
func jsonLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// logEntry represents a single JSON log line.
type logEntry map[string]any

// parseLogs parses a buffer of newline-delimited JSON log entries.
func parseLogs(t *testing.T, buf *bytes.Buffer) []logEntry {
	t.Helper()
	var entries []logEntry
	dec := json.NewDecoder(buf)
	for dec.More() {
		var entry logEntry
		if err := dec.Decode(&entry); err != nil {
			t.Fatalf("failed to parse log entry: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

// findOperation returns the first log entry whose "operation" field matches op.
func findOperation(entries []logEntry, op string) (logEntry, bool) {
	for _, e := range entries {
		if e["operation"] == op {
			return e, true
		}
	}
	return nil, false
}

// ──────────────────────────────────────────────────────────────────────────────
// Existing tests, updated to pass a logger
// ──────────────────────────────────────────────────────────────────────────────

func TestLocalScanRunner_Run_Success(t *testing.T) {
	tempDir := t.TempDir()

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), noop.New())

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
	if result.EffectiveConfig.MaxFileSizeBytes == 0 {
		t.Error("expected EffectiveConfig to be populated with defaults")
	}
}

func TestLocalScanRunner_Run_ResolveError(t *testing.T) {
	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), noop.New())

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
	code, ok := rcerr.CodeOf(err)
	if !ok {
		t.Fatalf("expected rcerr.Error in chain, got %T: %v", err, err)
	}
	if code != rcerr.CodeInvalidSource {
		t.Errorf("expected code %q, got %q", rcerr.CodeInvalidSource, code)
	}
}

func TestLocalScanRunner_Run_CancelledContextFails(t *testing.T) {
	tempDir := t.TempDir()

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), noop.New())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	result, err := runner.Run(ctx, req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}

	s := result.Scan
	if s.Status != scan.StatusFailed {
		t.Errorf("expected status %q, got %q", scan.StatusFailed, s.Status)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Log-capture tests: assert structured log fields
// ──────────────────────────────────────────────────────────────────────────────

func TestLocalScanRunner_Run_LogsSuccessLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	var buf bytes.Buffer

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, jsonLogger(&buf), noop.New())

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

	entries := parseLogs(t, &buf)

	// Assert operation sequence
	expectedOps := []string{"scan_start", "config_resolve", "repo_resolve", "snapshot_create", "scan_completed"}
	for _, op := range expectedOps {
		e, ok := findOperation(entries, op)
		if !ok {
			t.Errorf("expected log entry with operation=%q, not found in %d entries", op, len(entries))
			continue
		}
		// All entries must carry scan_id
		if e["scan_id"] == "" || e["scan_id"] == nil {
			t.Errorf("operation=%q: expected non-empty scan_id field", op)
		}
	}

	// repo_resolve must carry repository_id
	if e, ok := findOperation(entries, "repo_resolve"); ok {
		if e["repository_id"] == "" || e["repository_id"] == nil {
			t.Error("operation=repo_resolve: expected non-empty repository_id field")
		}
	}

	// snapshot_create must carry snapshot_id
	if e, ok := findOperation(entries, "snapshot_create"); ok {
		if e["snapshot_id"] == "" || e["snapshot_id"] == nil {
			t.Error("operation=snapshot_create: expected non-empty snapshot_id field")
		}
	}

	// scan_completed must carry repository_id, snapshot_id, and duration_ms
	if e, ok := findOperation(entries, "scan_completed"); ok {
		if e["repository_id"] == "" || e["repository_id"] == nil {
			t.Error("operation=scan_completed: expected non-empty repository_id field")
		}
		if e["snapshot_id"] == "" || e["snapshot_id"] == nil {
			t.Error("operation=scan_completed: expected non-empty snapshot_id field")
		}
		if _, hasDuration := e["duration_ms"]; !hasDuration {
			t.Error("operation=scan_completed: expected duration_ms field")
		}
	}

	// Assert scan IDs match between log and result
	if e, ok := findOperation(entries, "scan_start"); ok {
		if e["scan_id"] != result.Scan.ID {
			t.Errorf("scan_id mismatch: log has %q, result has %q", e["scan_id"], result.Scan.ID)
		}
	}
}

func TestLocalScanRunner_Run_LogsFailureWithErrorID(t *testing.T) {
	var buf bytes.Buffer

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, jsonLogger(&buf), noop.New())

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: filepath.Join(os.TempDir(), "rc_nonexistent_path_for_log_test"),
		},
	}

	_, err := runner.Run(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	entries := parseLogs(t, &buf)

	// scan_start must be present
	if _, ok := findOperation(entries, "scan_start"); !ok {
		t.Error("expected log entry with operation=scan_start")
	}

	// scan_failed must be present with error_id
	e, ok := findOperation(entries, "scan_failed")
	if !ok {
		t.Fatal("expected log entry with operation=scan_failed")
	}
	if e["error_id"] == "" || e["error_id"] == nil {
		t.Error("operation=scan_failed: expected non-empty error_id field")
	}
	if e["error_msg"] == "" || e["error_msg"] == nil {
		t.Error("operation=scan_failed: expected non-empty error_msg field")
	}
	if _, hasDuration := e["duration_ms"]; !hasDuration {
		t.Error("operation=scan_failed: expected duration_ms field")
	}
}
