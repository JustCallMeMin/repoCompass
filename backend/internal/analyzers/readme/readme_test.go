package readme_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/readme"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestAnalyzerMetadata(t *testing.T) {
	metadata := readme.New().Metadata()

	if metadata.ID != "readme" {
		t.Fatalf("unexpected analyzer ID: got %q", metadata.ID)
	}
	if metadata.Name != "README Analyzer" {
		t.Fatalf("unexpected analyzer name: got %q", metadata.Name)
	}
	if metadata.Version != "0.1.0" {
		t.Fatalf("unexpected analyzer version: got %q", metadata.Version)
	}
}

func TestAnalyzerReturnsSuccessWhenReadmeExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := readme.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}

	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(result.Findings))
	}
	if result.Metadata["readme_present"] != "true" {
		t.Fatalf("expected readme_present=true, got %q", result.Metadata["readme_present"])
	}
}

func TestAnalyzerReturnsFindingWhenReadmeMissing(t *testing.T) {
	dir := t.TempDir()

	result, err := readme.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}

	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Findings))
	}

	finding := result.Findings[0]
	if finding.RuleID != "readme.exists" {
		t.Fatalf("unexpected rule ID: got %q", finding.RuleID)
	}
	if finding.AnalyzerID != "readme" {
		t.Fatalf("unexpected analyzer ID: got %q", finding.AnalyzerID)
	}
	if finding.Severity != rules.SeverityHigh {
		t.Fatalf("unexpected severity: got %q", finding.Severity)
	}
	if finding.Category != rules.CategoryDocumentation {
		t.Fatalf("unexpected category: got %q", finding.Category)
	}
	if len(finding.Evidence) != 1 {
		t.Fatalf("expected one evidence item, got %d", len(finding.Evidence))
	}
	if finding.Evidence[0].Type != findings.EvidenceTypeFileMissing {
		t.Fatalf("unexpected evidence type: got %q", finding.Evidence[0].Type)
	}
	if len(finding.Recommendations) != 1 {
		t.Fatalf("expected one recommendation, got %d", len(finding.Recommendations))
	}
	if err := finding.Validate(); err != nil {
		t.Fatalf("expected finding to validate: %v", err)
	}
}

func TestAnalyzerSkipsWhenReadmeRuleDisabled(t *testing.T) {
	dir := t.TempDir()
	ruleSet, err := rules.ResolveRuleSet(rules.DefaultRuleSet(), configWithDisabledReadme())
	if err != nil {
		t.Fatalf("expected ruleset resolution to succeed: %v", err)
	}

	result, err := readme.New().Analyze(context.Background(), inputForPath(dir, ruleSet))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}

	if result.Status != analyzer.AnalyzerStatusSkipped {
		t.Fatalf("unexpected status: got %q", result.Status)
	}
	if result.Metadata["reason"] != "rule_disabled" {
		t.Fatalf("unexpected skip reason: got %q", result.Metadata["reason"])
	}
}

func TestAnalyzerReturnsErrorForInvalidRepositoryPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "repo-file")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readme.New().Analyze(context.Background(), inputForPath(filePath, rules.DefaultRuleSet()))
	if err == nil {
		t.Fatal("expected analyzer error for invalid repository path")
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

func configWithDisabledReadme() config.EffectiveConfiguration {
	return config.EffectiveConfiguration{DisabledRules: []string{"readme.exists"}}
}
