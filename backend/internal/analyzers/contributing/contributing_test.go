package contributing_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/contributing"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestAnalyzerMissingFileReturnsFinding(t *testing.T) {
	result, err := contributing.New().Analyze(context.Background(), inputForPath(t.TempDir(), rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	assertFinding(t, result, "contributing.exists")
}

func TestAnalyzerGoodFileReturnsNoFindings(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "CONTRIBUTING.md", "## Setup\nRun install.\n## Test\nRun go test ./...\n")

	result, err := contributing.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(result.Findings))
	}
}

func TestAnalyzerMissingSetupGuidanceReturnsFinding(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "CONTRIBUTING.md", "## Test\nRun go test ./...\n")

	result, err := contributing.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	assertFinding(t, result, "contributing.setup_instructions")
}

func TestAnalyzerMissingTestGuidanceReturnsFinding(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "CONTRIBUTING.md", "## Setup\nRun install.\n")

	result, err := contributing.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	assertFinding(t, result, "contributing.test_instructions")
}

func TestAnalyzerSkipsWhenRulesDisabled(t *testing.T) {
	ruleSet, err := rules.ResolveRuleSet(rules.DefaultRuleSet(), config.EffectiveConfiguration{
		DisabledRules: []string{"contributing.exists", "contributing.setup_instructions", "contributing.test_instructions"},
	})
	if err != nil {
		t.Fatalf("expected ruleset resolution to succeed: %v", err)
	}

	result, err := contributing.New().Analyze(context.Background(), inputForPath(t.TempDir(), ruleSet))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	if result.Status != analyzer.AnalyzerStatusSkipped {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
}

func TestAnalyzerInvalidRepoPathReturnsError(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "repo-file")
	writeFile(t, dir, "repo-file", "not a directory")

	_, err := contributing.New().Analyze(context.Background(), inputForPath(filePath, rules.DefaultRuleSet()))
	if err == nil {
		t.Fatal("expected analyzer error")
	}
}

func inputForPath(path string, ruleSet rules.RuleSet) analyzer.Input {
	return analyzer.Input{
		Repository: repository.Repository{
			ID:        "local_test",
			Name:      "test",
			LocalPath: path,
			Provider:  repository.ProviderLocal,
			Status:    repository.StatusActive,
		},
		RuleSet: ruleSet,
	}
}

func writeFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertFinding(t *testing.T, result analyzer.AnalyzerResult, ruleID string) {
	t.Helper()
	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	for _, finding := range result.Findings {
		if finding.RuleID == ruleID {
			if err := finding.Validate(); err != nil {
				t.Fatalf("expected finding to validate: %v", err)
			}
			return
		}
	}
	t.Fatalf("expected finding for rule %q, got %+v", ruleID, result.Findings)
}
