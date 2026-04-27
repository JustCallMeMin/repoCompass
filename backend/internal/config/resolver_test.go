package config

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
)

func TestLocalConfigurationResolver_ResolveConfig(t *testing.T) {
	resolver := NewLocalConfigurationResolver()
	ctx := context.Background()
	defaults := GetDefaults()

	t.Run("invalid context", func(t *testing.T) {
		_, err := resolver.ResolveConfig(nil, ".", Config{})
		if err == nil {
			t.Fatal("expected error for nil context")
		}
	})

	t.Run("invalid repoPath", func(t *testing.T) {
		_, err := resolver.ResolveConfig(ctx, "nonexistent_dir_12345", Config{})
		assertErrorCode(t, err, rcerr.CodeInvalidSource)
	})

	t.Run("repoPath is file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "somefile.txt")
		os.WriteFile(file, []byte("test"), 0644)
		_, err := resolver.ResolveConfig(ctx, file, Config{})
		assertErrorCode(t, err, rcerr.CodeInvalidSource)
	})

	t.Run("defaults only", func(t *testing.T) {
		dir := t.TempDir()
		
		got, err := resolver.ResolveConfig(ctx, dir, Config{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := Resolve([]Config{defaults})
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got %+v, want %+v", got, expected)
		}
	})

	t.Run("file override", func(t *testing.T) {
		dir := t.TempDir()
		cfgContent := []byte(`
excludes:
  - "override_dir"
maxFilesizeBytes: 5000000
enableDefaultAnalyzers: false
`)
		err := os.WriteFile(filepath.Join(dir, ".repocompass.yaml"), cfgContent, 0644)
		if err != nil {
			t.Fatal(err)
		}

		got, err := resolver.ResolveConfig(ctx, dir, Config{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got.Excludes) != 1 || got.Excludes[0] != "override_dir" {
			t.Errorf("unexpected excludes: %v", got.Excludes)
		}
		if got.MaxFileSizeBytes != 5000000 {
			t.Errorf("unexpected MaxFileSizeBytes: %d", got.MaxFileSizeBytes)
		}
		if got.EnableDefaultAnalyzers != false {
			t.Errorf("unexpected EnableDefaultAnalyzers: %v", got.EnableDefaultAnalyzers)
		}
	})

	t.Run("yml file override", func(t *testing.T) {
		dir := t.TempDir()
		cfgContent := []byte(`
excludes:
  - "yml_override"
maxFilesizeBytes: 7000000
enableDefaultAnalyzers: false
`)
		err := os.WriteFile(filepath.Join(dir, ".repocompass.yml"), cfgContent, 0644)
		if err != nil {
			t.Fatal(err)
		}

		got, err := resolver.ResolveConfig(ctx, dir, Config{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got.Excludes) != 1 || got.Excludes[0] != "yml_override" {
			t.Errorf("unexpected excludes: %v", got.Excludes)
		}
		if got.MaxFileSizeBytes != 7000000 {
			t.Errorf("unexpected MaxFileSizeBytes: %d", got.MaxFileSizeBytes)
		}
		if got.EnableDefaultAnalyzers != false {
			t.Errorf("unexpected EnableDefaultAnalyzers: %v", got.EnableDefaultAnalyzers)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, ".repocompass.yml"), []byte("invalid:\n  yaml:\n - structure"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = resolver.ResolveConfig(ctx, dir, Config{})
		assertErrorCode(t, err, rcerr.CodeConfigResolveFailed)
	})

	t.Run("CLI overrides file and defaults", func(t *testing.T) {
		dir := t.TempDir()
		cfgContent := []byte(`
excludes:
  - "override_dir"
maxFilesizeBytes: 5000000
`)
		os.WriteFile(filepath.Join(dir, ".repocompass.yaml"), cfgContent, 0644)

		cliMaxFileSize := int64(10)
		overrides := Config{
			MaxFileSizeBytes: &cliMaxFileSize,
		}

		got, err := resolver.ResolveConfig(ctx, dir, overrides)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// From file
		if len(got.Excludes) != 1 || got.Excludes[0] != "override_dir" {
			t.Errorf("unexpected excludes: %v", got.Excludes)
		}
		// From defaults
		if got.EnableDefaultAnalyzers != true {
			t.Errorf("unexpected EnableDefaultAnalyzers: %v", got.EnableDefaultAnalyzers)
		}
		// From overrides
		if got.MaxFileSizeBytes != 10 {
			t.Errorf("unexpected MaxFileSizeBytes: %d", got.MaxFileSizeBytes)
		}
	})

	t.Run("empty excludes replacement behavior", func(t *testing.T) {
		dir := t.TempDir()
		
		overrides := Config{
			Excludes: []string{}, // Empty slice to override defaults
		}

		got, err := resolver.ResolveConfig(ctx, dir, overrides)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got.Excludes) != 0 {
			t.Errorf("expected empty excludes, got: %v", got.Excludes)
		}
	})
}

func assertErrorCode(t *testing.T, err error, expected rcerr.ErrorCode) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error code %q, got nil", expected)
	}
	code, ok := rcerr.CodeOf(err)
	if !ok {
		t.Fatalf("expected rcerr.Error, got %T: %v", err, err)
	}
	if code != expected {
		t.Errorf("expected code %q, got %q", expected, code)
	}
}
