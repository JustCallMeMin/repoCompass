package example

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
)

// ExampleAnalyzer is a minimal example of an analyzer for new contributors to learn from.
type ExampleAnalyzer struct{}

// Metadata returns the identity and version of the analyzer.
func (ExampleAnalyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{
		ID:      "example",
		Name:    "Example Analyzer",
		Version: "0.1.0",
	}
}

// Analyze inspects the repository snapshot and returns the results.
func (ExampleAnalyzer) Analyze(ctx context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	// A real analyzer would use input.Snapshot.ReadFile() to inspect files.

	// Return a successful result with no findings.
	return analyzer.AnalyzerResult{
		AnalyzerID: "example",
		Name:       "Example Analyzer",
		Version:    "0.1.0",
		Status:     analyzer.AnalyzerStatusSuccess,
		Findings:   []findings.Finding{}, // Empty means no issues found
	}, nil
}
