package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestScanCmd_FixtureHappyPath(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "basic-go-repo")

	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan", fixturePath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := stdout.String()
	assertScanContains(t, output, "Starting scan for repository at")
	assertScanContains(t, output, "Scan Summary:")
	assertScanContains(t, output, "Status:       completed")
	assertScanContains(t, output, "Analyzers:    4")
	assertScanContains(t, output, "Findings:")
	assertScanContains(t, output, "Score:")
	assertScanContains(t, output, "Max File Size: 1048576 bytes")
	assertMatches(t, output, `Scan ID:\s+scan_[[:xdigit:]]+`)
	assertMatches(t, output, `Snapshot ID:\s+snap_[[:xdigit:]]+`)

	logs := stderr.String()
	assertScanContains(t, logs, "operation=scan_start")
	assertScanContains(t, logs, "operation=scan_completed")
	assertScanContains(t, logs, "repository_id=local_")
	assertScanContains(t, logs, "snapshot_id=snap_")
}

func TestScanCmd_FormatMarkdownWritesReportToStdout(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "missing-readme-repo")

	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan", fixturePath, "--format", "markdown"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := stdout.String()
	assertScanContains(t, output, "# RepoCompass Report")
	assertScanContains(t, output, "## Findings")
	assertScanContains(t, output, "README file is missing")
	assertScanNotContains(t, output, "Starting scan for repository at")
	assertScanContains(t, stderr.String(), "operation=scan_completed")
}

func TestScanCmd_FormatJSONWritesReportToStdout(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "missing-readme-repo")

	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan", fixturePath, "--format", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := stdout.String()
	assertScanContains(t, output, `"schema_version": "0.1.0"`)
	assertScanContains(t, output, `"findings":`)
	assertScanContains(t, output, `"rule_id": "readme.exists"`)
	assertScanNotContains(t, output, "Starting scan for repository at")
}

func TestScanCmd_OutputWritesReportFile(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "missing-readme-repo")
	outputPath := filepath.Join(t.TempDir(), "report.md")

	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan", fixturePath, "--format", "markdown", "--output", outputPath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stdout.String() != "" {
		t.Fatalf("expected stdout to be empty when writing output file, got %q", stdout.String())
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected report file to be written: %v", err)
	}
	assertScanContains(t, string(content), "# RepoCompass Report")
	assertScanContains(t, string(content), "README file is missing")
}

func TestScanCmd_InvalidFormatFails(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "basic-go-repo")

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"scan", fixturePath, "--format", "xml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected invalid format to fail")
	}
	assertScanContains(t, err.Error(), "unsupported report format")
}

func TestScanCmd_OutputWithoutFormatFails(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", "basic-go-repo")

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"scan", fixturePath, "--output", filepath.Join(t.TempDir(), "report.md")})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected --output without --format to fail")
	}
	assertScanContains(t, err.Error(), "--output requires --format")
}

func TestScanCmd_FixtureInvalidPath(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing-repository")

	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan", missingPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	assertScanContains(t, err.Error(), "scan failed [INVALID_SOURCE]")
	assertScanContains(t, stdout.String(), "Starting scan for repository at")
	assertScanNotContains(t, stdout.String(), "Scan Summary:")
	assertScanContains(t, stderr.String(), "operation=scan_failed")
	assertScanContains(t, stderr.String(), "error_id=INVALID_SOURCE")
}

func TestScanCmd_Failure_MissingArgument(t *testing.T) {
	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"scan"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "accepts 1 arg(s), received 0") {
		t.Errorf("expected missing argument error, got %v", err)
	}
}

func assertScanContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

func assertScanNotContains(t *testing.T, got, unwanted string) {
	t.Helper()
	if strings.Contains(got, unwanted) {
		t.Fatalf("expected %q not to contain %q", got, unwanted)
	}
}

func assertMatches(t *testing.T, got, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(got) {
		t.Fatalf("expected %q to match %q", got, pattern)
	}
}
