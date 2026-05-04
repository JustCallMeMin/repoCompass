package analyzer_test

import (
	"context"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestFakeAnalyzerImplementsContract(t *testing.T) {
	var _ analyzer.Analyzer = fakeAnalyzer{}
}

func TestAnalyzerReceivesInputAndReturnsResult(t *testing.T) {
	a := fakeAnalyzer{}
	input := analyzer.Input{
		Repository: repository.Repository{
			ID:       "local_repo",
			Name:     "example",
			Provider: repository.ProviderLocal,
		},
		Snapshot: snapshot.RepositorySnapshot{
			ID:           "snap_123",
			RepositoryID: "local_repo",
			SourceType:   snapshot.SourceTypeLocal,
		},
		EffectiveConfiguration: config.EffectiveConfiguration{
			MaxFileSizeBytes:       1024,
			EnableDefaultAnalyzers: true,
		},
		RuleSet: rules.DefaultRuleSet(),
	}

	result, err := a.Analyze(context.Background(), input)
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}

	if result.AnalyzerID != "fake" {
		t.Fatalf("unexpected analyzer ID: got %q", result.AnalyzerID)
	}
	if result.Name != "Fake Analyzer" {
		t.Fatalf("unexpected analyzer name: got %q", result.Name)
	}
	if result.Version != "0.1.0" {
		t.Fatalf("unexpected analyzer version: got %q", result.Version)
	}
	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Fatalf("unexpected analyzer status: got %q", result.Status)
	}
	if result.Duration != 5*time.Millisecond {
		t.Fatalf("unexpected analyzer duration: got %s", result.Duration)
	}
	if result.Metadata["repository_id"] != input.Repository.ID {
		t.Fatalf("expected repository ID metadata %q, got %q", input.Repository.ID, result.Metadata["repository_id"])
	}
	if result.Metadata["snapshot_id"] != input.Snapshot.ID {
		t.Fatalf("expected snapshot ID metadata %q, got %q", input.Snapshot.ID, result.Metadata["snapshot_id"])
	}
	if result.Metadata["default_analyzers"] != "true" {
		t.Fatalf("expected config metadata to be carried, got %q", result.Metadata["default_analyzers"])
	}
	if result.Metadata["ruleset_id"] != input.RuleSet.ID {
		t.Fatalf("expected ruleset metadata %q, got %q", input.RuleSet.ID, result.Metadata["ruleset_id"])
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Findings))
	}
	if result.Findings[0].RuleID != "fake.rule" {
		t.Fatalf("unexpected finding rule ID: got %q", result.Findings[0].RuleID)
	}
}

func TestAnalyzerResultSupportsSkippedStatus(t *testing.T) {
	result := analyzer.AnalyzerResult{
		AnalyzerID: "fake",
		Name:       "Fake Analyzer",
		Version:    "0.1.0",
		Status:     analyzer.AnalyzerStatusSkipped,
		Metadata: map[string]string{
			"reason": "disabled",
		},
	}

	if result.Status != analyzer.AnalyzerStatusSkipped {
		t.Fatalf("unexpected analyzer status: got %q", result.Status)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings for skipped result, got %d", len(result.Findings))
	}
	if result.Metadata["reason"] != "disabled" {
		t.Fatalf("unexpected skipped reason: got %q", result.Metadata["reason"])
	}
}

func TestAnalyzerResultSupportsFailedStatus(t *testing.T) {
	result := analyzer.AnalyzerResult{
		AnalyzerID:   "fake",
		Name:         "Fake Analyzer",
		Version:      "0.1.0",
		Status:       analyzer.AnalyzerStatusFailed,
		ErrorMessage: "failed to read repository file",
	}

	if result.Status != analyzer.AnalyzerStatusFailed {
		t.Fatalf("unexpected analyzer status: got %q", result.Status)
	}
	if result.ErrorMessage == "" {
		t.Fatal("expected failed result to carry an error message")
	}
}

func TestAnalyzerMetadata(t *testing.T) {
	metadata := fakeAnalyzer{}.Metadata()

	if metadata.ID != "fake" {
		t.Fatalf("unexpected metadata ID: got %q", metadata.ID)
	}
	if metadata.Name != "Fake Analyzer" {
		t.Fatalf("unexpected metadata name: got %q", metadata.Name)
	}
	if metadata.Version != "0.1.0" {
		t.Fatalf("unexpected metadata version: got %q", metadata.Version)
	}
}

type fakeAnalyzer struct{}

func (fakeAnalyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{
		ID:      "fake",
		Name:    "Fake Analyzer",
		Version: "0.1.0",
	}
}

func (fakeAnalyzer) Analyze(_ context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	return analyzer.AnalyzerResult{
		AnalyzerID: "fake",
		Name:       "Fake Analyzer",
		Version:    "0.1.0",
		Status:     analyzer.AnalyzerStatusSuccess,
		Duration:   5 * time.Millisecond,
		Metadata: map[string]string{
			"repository_id":     input.Repository.ID,
			"snapshot_id":       input.Snapshot.ID,
			"default_analyzers": boolString(input.EffectiveConfiguration.EnableDefaultAnalyzers),
			"ruleset_id":        input.RuleSet.ID,
		},
		Findings: []findings.Finding{
			findings.NewFinding(
				"finding_fake",
				"fake.rule",
				"fake",
				rules.SeverityLow,
				rules.CategoryMaintainability,
				"Fake finding",
				"Fake analyzer received input and returned a finding.",
				[]findings.Evidence{
					findings.NewEvidence(
						findings.EvidenceTypeMetadata,
						"Fake analyzer received repository and snapshot metadata.",
						"",
						"repository_id="+input.Repository.ID,
					),
				},
			),
		},
	}, nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
