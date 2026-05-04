package analyzers_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/ci"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/contributing"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/readme"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/scripts"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestAnalyzerFixtures(t *testing.T) {
	tests := []struct {
		name        string
		fixtureName string
		wantRules   []string
	}{
		{
			name:        "good fixture produces no findings",
			fixtureName: "good-onboarding-repo",
			wantRules:   nil,
		},
		{
			name:        "missing README fixture produces README finding",
			fixtureName: "missing-readme-repo",
			wantRules:   []string{"readme.exists"},
		},
		{
			name:        "missing CONTRIBUTING fixture produces CONTRIBUTING finding",
			fixtureName: "missing-contributing-repo",
			wantRules:   []string{"contributing.exists"},
		},
		{
			name:        "missing CI fixture produces CI finding",
			fixtureName: "missing-ci-repo",
			wantRules:   []string{"ci.workflow.exists"},
		},
		{
			name:        "missing scripts fixture produces scripts finding",
			fixtureName: "missing-scripts-repo",
			wantRules:   []string{"scripts.dev.exists"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := runAnalyzers(t, fixturePath(tt.fixtureName))
			gotRules := findingRuleIDs(results)
			if len(gotRules) != len(tt.wantRules) {
				t.Fatalf("expected finding rules %v, got %v", tt.wantRules, gotRules)
			}
			for _, want := range tt.wantRules {
				if !contains(gotRules, want) {
					t.Fatalf("expected finding rule %q, got %v", want, gotRules)
				}
			}
		})
	}
}

func runAnalyzers(t *testing.T, repoPath string) []analyzer.AnalyzerResult {
	t.Helper()
	input := analyzer.Input{
		Repository: repository.Repository{
			ID:        "fixture",
			Name:      filepath.Base(repoPath),
			LocalPath: repoPath,
			Provider:  repository.ProviderLocal,
			Status:    repository.StatusActive,
		},
		RuleSet: rules.DefaultRuleSet(),
	}
	analyzers := []analyzer.Analyzer{
		readme.New(),
		contributing.New(),
		ci.New(),
		scripts.New(),
	}
	results := make([]analyzer.AnalyzerResult, 0, len(analyzers))
	for _, current := range analyzers {
		result, err := current.Analyze(context.Background(), input)
		if err != nil {
			t.Fatalf("analyzer %s failed: %v", current.Metadata().ID, err)
		}
		results = append(results, result)
	}
	return results
}

func fixturePath(name string) string {
	return filepath.Join("..", "..", "testdata", "fixtures", "local-repositories", name)
}

func findingRuleIDs(results []analyzer.AnalyzerResult) []string {
	var ids []string
	for _, result := range results {
		for _, finding := range result.Findings {
			ids = append(ids, finding.RuleID)
		}
	}
	return ids
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
