package scan_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// fakeStore records all calls made by the runner so tests can assert
// that persistence methods are called in the correct order.
type fakeStore struct {
	mu   sync.Mutex
	calls []string
	failOn string // if non-empty, return an error when this method is called
}

func newFakeStore() *fakeStore { return &fakeStore{} }

func (f *fakeStore) record(method string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, method)
	if f.failOn == method {
		return errors.New("fakeStore: injected failure for " + method)
	}
	return nil
}

func (f *fakeStore) SaveRepository(_ context.Context, _ repository.Repository) error {
	return f.record("SaveRepository")
}
func (f *fakeStore) SaveSnapshot(_ context.Context, _ snapshot.RepositorySnapshot) error {
	return f.record("SaveSnapshot")
}
func (f *fakeStore) SaveScan(_ context.Context, _ scan.Scan) error {
	return f.record("SaveScan")
}
func (f *fakeStore) UpdateScan(_ context.Context, _ scan.Scan) error {
	return f.record("UpdateScan")
}

func (f *fakeStore) assertCallOrder(t *testing.T, expected []string) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.calls) != len(expected) {
		t.Errorf("store call count: want %d, got %d; calls: %v", len(expected), len(f.calls), f.calls)
		return
	}
	for i, want := range expected {
		if f.calls[i] != want {
			t.Errorf("store call[%d]: want %q, got %q", i, want, f.calls[i])
		}
	}
}

// TestLocalScanRunner_PersistsInOrder verifies the runner calls the store
// in the correct lifecycle sequence on a successful scan.
func TestLocalScanRunner_PersistsInOrder(t *testing.T) {
	tempDir := t.TempDir()
	store := newFakeStore()

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), store)

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	_, err := runner.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	store.assertCallOrder(t, []string{
		"SaveRepository",
		"SaveSnapshot",
		"SaveScan",
		"UpdateScan",
	})
}

// TestLocalScanRunner_StoreFailurePropagatesToScan verifies that a store failure
// causes the scan to fail with a non-zero error.
func TestLocalScanRunner_SaveRepositoryFailureFails(t *testing.T) {
	tempDir := t.TempDir()
	store := newFakeStore()
	store.failOn = "SaveRepository"

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), store)

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	result, err := runner.Run(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from store failure, got nil")
	}
	if result.Scan.Status != scan.StatusFailed {
		t.Errorf("expected scan status %q, got %q", scan.StatusFailed, result.Scan.Status)
	}
}

// TestLocalScanRunner_UpdateScanFailurePropagates verifies that a failure in
// UpdateScan (the final step) surfaces as an error even when scan execution succeeded.
func TestLocalScanRunner_UpdateScanFailurePropagates(t *testing.T) {
	tempDir := t.TempDir()
	store := newFakeStore()
	store.failOn = "UpdateScan"

	provider := repository.NewLocalRepositoryProvider()
	creator := snapshot.NewCreator()
	resolver := config.NewLocalConfigurationResolver()
	runner := scan.NewLocalScanRunner(provider, creator, resolver, discardLogger(), store)

	req := scan.RunRequest{
		Source: repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: tempDir,
		},
	}

	_, err := runner.Run(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from UpdateScan failure, got nil")
	}
}
