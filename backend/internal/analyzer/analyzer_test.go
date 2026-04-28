package analyzer_test

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
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
	}

	result, err := a.Analyze(context.Background(), input)
	if err != nil {
		t.Fatalf("expected analyze to succeed: %v", err)
	}

	if result.AnalyzerID != "fake" {
		t.Fatalf("unexpected analyzer ID: got %q", result.AnalyzerID)
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
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Findings))
	}
	if result.Findings[0].RuleID != "fake.rule" {
		t.Fatalf("unexpected finding rule ID: got %q", result.Findings[0].RuleID)
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

func (fakeAnalyzer) Metadata() analyzer.Metadata {
	return analyzer.Metadata{
		ID:      "fake",
		Name:    "Fake Analyzer",
		Version: "0.1.0",
	}
}

func (fakeAnalyzer) Analyze(_ context.Context, input analyzer.Input) (analyzer.Result, error) {
	return analyzer.Result{
		AnalyzerID: "fake",
		Metadata: map[string]string{
			"repository_id":     input.Repository.ID,
			"snapshot_id":       input.Snapshot.ID,
			"default_analyzers": boolString(input.EffectiveConfiguration.EnableDefaultAnalyzers),
		},
		Findings: []analyzer.Finding{
			{
				RuleID:  "fake.rule",
				Title:   "Fake finding",
				Message: "Fake analyzer received input and returned a finding.",
			},
		},
	}, nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
