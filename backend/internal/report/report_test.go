package report_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/report"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestFakeRendererImplementsRenderer(t *testing.T) {
	var _ report.Renderer = fakeRenderer{}
}

func TestFormatConstantsAreStable(t *testing.T) {
	if report.FormatMarkdown != "markdown" {
		t.Fatalf("unexpected markdown format: %q", report.FormatMarkdown)
	}
	if report.FormatJSON != "json" {
		t.Fatalf("unexpected json format: %q", report.FormatJSON)
	}
}

func TestRenderRequestCarriesScanContext(t *testing.T) {
	request := report.RenderRequest{
		Scan:            scan.Scan{ID: "scan_123"},
		Repository:      repository.Repository{ID: "repo_123"},
		Snapshot:        snapshot.RepositorySnapshot{ID: "snap_123"},
		AnalyzerResults: []analyzer.AnalyzerResult{{AnalyzerID: "readme"}},
		Assessment:      assessment.Assessment{OverallScore: 100},
		EffectiveConfig: config.EffectiveConfiguration{MaxFileSizeBytes: 1024},
	}

	result, err := fakeRenderer{}.Render(context.Background(), request)
	if err != nil {
		t.Fatalf("expected render to succeed: %v", err)
	}
	if result.Format != report.FormatMarkdown {
		t.Fatalf("unexpected render format: %q", result.Format)
	}
	if string(result.Content) != "scan_123 repo_123 snap_123 readme 100 1024" {
		t.Fatalf("unexpected render content: %q", string(result.Content))
	}
	if result.ContentType != "text/markdown" {
		t.Fatalf("unexpected content type: %q", result.ContentType)
	}
}

type fakeRenderer struct{}

func (fakeRenderer) Format() report.Format {
	return report.FormatMarkdown
}

func (fakeRenderer) Render(_ context.Context, request report.RenderRequest) (report.RenderResult, error) {
	content := request.Scan.ID + " " +
		request.Repository.ID + " " +
		request.Snapshot.ID + " " +
		request.AnalyzerResults[0].AnalyzerID + " " +
		strconv.Itoa(request.Assessment.OverallScore) + " " +
		strconv.FormatInt(request.EffectiveConfig.MaxFileSizeBytes, 10)
	return report.RenderResult{
		Format:      report.FormatMarkdown,
		Content:     []byte(content),
		ContentType: "text/markdown",
	}, nil
}
