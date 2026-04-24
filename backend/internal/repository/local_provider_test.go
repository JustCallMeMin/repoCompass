package repository

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalRepositoryProviderResolvesValidPath(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "basic-go-repo")
	absolutePath, err := filepath.Abs(fixturePath)
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}

	provider := LocalRepositoryProvider{
		now: func() time.Time {
			return time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
		},
	}

	resolution, err := provider.Resolve(context.Background(), RepositorySource{
		Type: SourceTypeLocal,
		Path: fixturePath,
	})
	if err != nil {
		t.Fatalf("expected local repository resolution to succeed: %v", err)
	}

	repo := resolution.Repository
	if repo.ID == "" {
		t.Fatal("expected repository ID to be set")
	}
	if repo.Name != "basic-go-repo" {
		t.Fatalf("unexpected repository name: got %q", repo.Name)
	}
	if repo.FullName != "basic-go-repo" {
		t.Fatalf("unexpected repository full name: got %q", repo.FullName)
	}
	if repo.Provider != ProviderLocal {
		t.Fatalf("unexpected provider: got %q", repo.Provider)
	}
	if repo.Status != StatusActive {
		t.Fatalf("unexpected status: got %q", repo.Status)
	}

	expectedURL := (&url.URL{Scheme: "file", Path: absolutePath}).String()
	if repo.URL != expectedURL {
		t.Fatalf("unexpected repository URL: got %q want %q", repo.URL, expectedURL)
	}

	metadata := resolution.SnapshotMetadata
	if metadata["local_path"] != absolutePath {
		t.Fatalf("unexpected metadata local_path: got %q want %q", metadata["local_path"], absolutePath)
	}
	if metadata["source_type"] != string(SourceTypeLocal) {
		t.Fatalf("unexpected metadata source_type: got %q", metadata["source_type"])
	}
	if metadata["resolved_at"] != "2026-04-24T12:00:00Z" {
		t.Fatalf("unexpected metadata resolved_at: got %q", metadata["resolved_at"])
	}
}

func TestLocalRepositoryProviderRejectsMissingPath(t *testing.T) {
	provider := NewLocalRepositoryProvider()

	_, err := provider.Resolve(context.Background(), RepositorySource{
		Type: SourceTypeLocal,
		Path: filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil {
		t.Fatal("expected missing path to fail")
	}
}

func TestLocalRepositoryProviderRejectsFilePath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "not-a-directory.txt")
	if err := os.WriteFile(filePath, []byte("not a repo"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}

	provider := NewLocalRepositoryProvider()
	_, err := provider.Resolve(context.Background(), RepositorySource{
		Type: SourceTypeLocal,
		Path: filePath,
	})
	if err == nil {
		t.Fatal("expected file path to fail")
	}
}

func TestLocalRepositoryProviderReportsSourceType(t *testing.T) {
	provider := NewLocalRepositoryProvider()

	if got := provider.SourceType(); got != SourceTypeLocal {
		t.Fatalf("unexpected source type: got %q want %q", got, SourceTypeLocal)
	}
}

func TestProviderRegistrySelectsLocalRepositoryProvider(t *testing.T) {
	localProvider := NewLocalRepositoryProvider()
	registry, err := NewProviderRegistry(localProvider)
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	provider, err := registry.ProviderFor(SourceTypeLocal)
	if err != nil {
		t.Fatalf("expected provider lookup to succeed: %v", err)
	}
	if provider.SourceType() != SourceTypeLocal {
		t.Fatalf("unexpected provider source type: got %q", provider.SourceType())
	}
}
