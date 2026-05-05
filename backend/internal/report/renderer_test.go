package report

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestMarkdownRendererRendersSectionsAndContent(t *testing.T) {
	renderer := MarkdownRenderer{now: fixedNow}

	result, err := renderer.Render(context.Background(), sampleRenderRequest())
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}

	if result.Format != FormatMarkdown {
		t.Fatalf("unexpected format: %q", result.Format)
	}
	if result.ContentType != "text/markdown" {
		t.Fatalf("unexpected content type: %q", result.ContentType)
	}
	output := string(result.Content)
	for _, want := range []string{
		"# RepoCompass Report",
		"## Repository Summary",
		"## Scan Summary",
		"## Assessment",
		"## Analyzer Results",
		"## Findings",
		"## Recommendations",
		"## Metadata",
		"README file is missing",
		"file_missing",
		"Add a root README",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected markdown output to contain %q\n%s", want, output)
		}
	}
}

func TestMarkdownRendererNoFindingsEmptyState(t *testing.T) {
	request := sampleRenderRequest()
	request.AnalyzerResults = []analyzer.AnalyzerResult{{AnalyzerID: "readme", Name: "README Analyzer", Version: "0.1.0", Status: analyzer.AnalyzerStatusSuccess}}
	request.Assessment = assessment.Assessment{OverallScore: 100, Label: assessment.ScoreLabelExcellent}

	result, err := (MarkdownRenderer{now: fixedNow}).Render(context.Background(), request)
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}
	output := string(result.Content)
	if !strings.Contains(output, "No findings detected.") {
		t.Fatalf("expected no findings empty state\n%s", output)
	}
	if !strings.Contains(output, "No recommendations.") {
		t.Fatalf("expected no recommendations empty state\n%s", output)
	}
}

func TestJSONRendererRendersStablePayload(t *testing.T) {
	renderer := JSONRenderer{now: fixedNow}

	result, err := renderer.Render(context.Background(), sampleRenderRequest())
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}
	if result.Format != FormatJSON {
		t.Fatalf("unexpected format: %q", result.Format)
	}
	if result.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", result.ContentType)
	}

	var payload map[string]any
	if err := json.Unmarshal(result.Content, &payload); err != nil {
		t.Fatalf("expected valid JSON: %v\n%s", err, string(result.Content))
	}
	for _, key := range []string{"schema_version", "repository", "scan", "assessment", "analyzers", "findings", "recommendations", "metadata"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected top-level key %q in payload %v", key, payload)
		}
	}
	if payload["schema_version"] != "0.1.0" {
		t.Fatalf("unexpected schema version: %v", payload["schema_version"])
	}

	metadata := payload["metadata"].(map[string]any)
	if metadata["generated_at"] != "2026-05-01T01:02:03Z" {
		t.Fatalf("unexpected generated_at: %v", metadata["generated_at"])
	}

	recommendations := payload["recommendations"].([]any)
	if len(recommendations) != 1 {
		t.Fatalf("expected one recommendation, got %d", len(recommendations))
	}
	recommendation := recommendations[0].(map[string]any)
	if recommendation["finding_id"] != "finding_readme" {
		t.Fatalf("unexpected recommendation finding_id: %v", recommendation["finding_id"])
	}
}

func TestMarkdownRendererMatchesGoldenFile(t *testing.T) {
	result, err := (MarkdownRenderer{now: fixedNow}).Render(context.Background(), sampleRenderRequest())
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}

	assertGolden(t, "markdown_report.md", result.Content)
}

func TestJSONRendererMatchesGoldenFile(t *testing.T) {
	result, err := (JSONRenderer{now: fixedNow}).Render(context.Background(), sampleRenderRequest())
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}

	assertGolden(t, "json_report.json", result.Content)
}

func sampleRenderRequest() RenderRequest {
	start := time.Date(2026, 5, 1, 1, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 1, 1, 0, 2, 0, time.UTC)
	finding := sampleFinding()
	return RenderRequest{
		Scan: scan.Scan{
			ID:         "scan_123",
			SnapshotID: "snap_123",
			Status:     scan.StatusCompleted,
			StartTime:  &start,
			EndTime:    &end,
		},
		Repository: repository.Repository{
			ID:            "repo_123",
			Name:          "example",
			URL:           "file:///repo/example",
			Provider:      repository.ProviderLocal,
			DefaultBranch: "main",
			LocalPath:     "/absolute/path/not/rendered/in/json",
		},
		Snapshot: snapshot.RepositorySnapshot{ID: "snap_123", RepositoryID: "repo_123"},
		AnalyzerResults: []analyzer.AnalyzerResult{
			{
				AnalyzerID: "readme",
				Name:       "README Analyzer",
				Version:    "0.1.0",
				Status:     analyzer.AnalyzerStatusSuccess,
				Findings:   []findings.Finding{finding},
			},
		},
		Assessment: assessment.Assessment{
			OverallScore:   75,
			Label:          assessment.ScoreLabelGood,
			FindingCount:   1,
			SeverityCounts: map[rules.Severity]int{rules.SeverityHigh: 1},
			CategoryScores: map[rules.Category]int{rules.CategoryDocumentation: 75},
		},
	}
}

func sampleFinding() findings.Finding {
	finding := findings.NewFinding(
		"finding_readme",
		"readme.exists",
		"readme",
		rules.SeverityHigh,
		rules.CategoryDocumentation,
		"README file is missing",
		"The repository does not contain a README file at its root.",
		[]findings.Evidence{
			findings.FileMissingEvidence("README.md", "README.md was not found at the repository root."),
		},
	)
	finding.Recommendations = []findings.Recommendation{
		findings.RecommendationForFinding(
			finding.ID,
			"Add a root README",
			"Create README.md with project purpose, setup steps, and test commands.",
			"New contributors need one stable entry point before changing code.",
			findings.RecommendationPriorityHigh,
		),
	}
	return finding
}

func fixedNow() time.Time {
	return time.Date(2026, 5, 1, 1, 2, 3, 0, time.UTC)
}

func assertGolden(t *testing.T, filename string, got []byte) {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "golden", "reports", filename)
	
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create golden dir: %v", err)
		}
		if err := os.WriteFile(path, got, 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}
	
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file %s: %v", path, err)
	}
	gotText := strings.ReplaceAll(string(got), "\r\n", "\n")
	wantText := strings.ReplaceAll(string(want), "\r\n", "\n")
	if gotText != wantText {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", filename, gotText, wantText)
	}
}
