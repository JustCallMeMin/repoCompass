package ci_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/ci"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestAnalyzerMissingWorkflowReturnsFinding(t *testing.T) {
	result, err := ci.New().Analyze(context.Background(), inputForPath(t.TempDir(), rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	assertFinding(t, result, "ci.workflow.exists")
}

func TestAnalyzerWorkflowWithTestSignalReturnsNoFindings(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "test.yml", "name: test\njobs:\n  test:\n    steps:\n      - run: go test ./...\n")

	result, err := ci.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
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

func TestAnalyzerWorkflowWithoutTestSignalReturnsFinding(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "build.yml", "name: build\njobs:\n  build:\n    steps:\n      - run: echo build\n")

	result, err := ci.New().Analyze(context.Background(), inputForPath(dir, rules.DefaultRuleSet()))
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}
	assertFinding(t, result, "ci.test_job.exists")
}

func TestAnalyzerSkipsWhenRulesDisabled(t *testing.T) {
	ruleSet, err := rules.ResolveRuleSet(rules.DefaultRuleSet(), config.EffectiveConfiguration{
		DisabledRules: []string{"ci.workflow.exists", "ci.test_job.exists"},
	})
	if err != nil {
		t.Fatalf("expected ruleset resolution to succeed: %v", err)
	}

	result, err := ci.New().Analyze(context.Background(), inputForPath(t.TempDir(), ruleSet))
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
	if err := os.WriteFile(filePath, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ci.New().Analyze(context.Background(), inputForPath(filePath, rules.DefaultRuleSet()))
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

func writeWorkflow(t *testing.T, dir string, name string, content string) {
	t.Helper()
	workflowDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workflowDir, name), []byte(content), 0644); err != nil {
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
