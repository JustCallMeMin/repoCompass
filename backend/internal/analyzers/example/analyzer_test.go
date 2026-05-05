package example

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
)

func TestExampleAnalyzer_Metadata(t *testing.T) {
	a := ExampleAnalyzer{}
	m := a.Metadata()
	if m.ID != "example" {
		t.Errorf("expected ID 'example', got %q", m.ID)
	}
}

func TestExampleAnalyzer_Analyze(t *testing.T) {
	a := ExampleAnalyzer{}
	result, err := a.Analyze(context.Background(), analyzer.Input{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != analyzer.AnalyzerStatusSuccess {
		t.Errorf("expected success status, got %q", result.Status)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}
