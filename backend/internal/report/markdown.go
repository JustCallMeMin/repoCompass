package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
)

// MarkdownRenderer renders human-readable Markdown reports.
type MarkdownRenderer struct {
	now func() time.Time
}

// NewMarkdownRenderer creates a Markdown renderer.
func NewMarkdownRenderer() Renderer {
	return MarkdownRenderer{now: time.Now}
}

// Format returns the Markdown report format ID.
func (r MarkdownRenderer) Format() Format {
	return FormatMarkdown
}

// Render converts scan result data into a Markdown report.
func (r MarkdownRenderer) Render(ctx context.Context, request RenderRequest) (RenderResult, error) {
	if err := ctx.Err(); err != nil {
		return RenderResult{}, err
	}
	now := r.now
	if now == nil {
		now = time.Now
	}

	findingsList := collectFindings(request.AnalyzerResults)
	recommendations := collectRecommendations(findingsList)

	var b strings.Builder
	b.WriteString("# RepoCompass Report\n\n")
	renderRepositorySummary(&b, request)
	renderScanSummary(&b, request, len(findingsList))
	renderAssessment(&b, request)
	renderAnalyzerResults(&b, request)
	renderFindings(&b, findingsList)
	renderRecommendations(&b, recommendations)
	renderMetadata(&b, string(FormatMarkdown), now().UTC())

	return RenderResult{
		Format:      FormatMarkdown,
		Content:     []byte(b.String()),
		ContentType: "text/markdown",
	}, nil
}

func renderRepositorySummary(b *strings.Builder, request RenderRequest) {
	b.WriteString("## Repository Summary\n\n")
	b.WriteString("| Field | Value |\n| --- | --- |\n")
	fmt.Fprintf(b, "| ID | %s |\n", markdownCell(request.Repository.ID))
	fmt.Fprintf(b, "| Name | %s |\n", markdownCell(request.Repository.Name))
	fmt.Fprintf(b, "| Provider | %s |\n", markdownCell(string(request.Repository.Provider)))
	fmt.Fprintf(b, "| URL | %s |\n", markdownCell(request.Repository.URL))
	fmt.Fprintf(b, "| Default Branch | %s |\n\n", markdownCell(request.Repository.DefaultBranch))
}

func renderScanSummary(b *strings.Builder, request RenderRequest, findingCount int) {
	b.WriteString("## Scan Summary\n\n")
	b.WriteString("| Field | Value |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Scan ID | %s |\n", markdownCell(request.Scan.ID))
	fmt.Fprintf(b, "| Snapshot ID | %s |\n", markdownCell(request.Scan.SnapshotID))
	fmt.Fprintf(b, "| Status | %s |\n", markdownCell(string(request.Scan.Status)))
	fmt.Fprintf(b, "| Started At | %s |\n", markdownCell(formatTimePointer(request.Scan.StartTime)))
	fmt.Fprintf(b, "| Completed At | %s |\n", markdownCell(formatTimePointer(request.Scan.EndTime)))
	fmt.Fprintf(b, "| Analyzers | %d |\n", len(request.AnalyzerResults))
	fmt.Fprintf(b, "| Findings | %d |\n\n", findingCount)
}

func renderAssessment(b *strings.Builder, request RenderRequest) {
	b.WriteString("## Assessment\n\n")
	fmt.Fprintf(b, "Overall score: **%d/100** (%s)\n\n", request.Assessment.OverallScore, request.Assessment.Label)
	if len(request.Assessment.SeverityCounts) > 0 {
		b.WriteString("### Severity Counts\n\n")
		b.WriteString("| Severity | Count |\n| --- | --- |\n")
		for _, severity := range sortedSeverityKeys(request.Assessment.SeverityCounts) {
			fmt.Fprintf(b, "| %s | %d |\n", severity, request.Assessment.SeverityCounts[severity])
		}
		b.WriteString("\n")
	}
	if len(request.Assessment.CategoryScores) > 0 {
		b.WriteString("### Category Scores\n\n")
		b.WriteString("| Category | Score |\n| --- | --- |\n")
		for _, category := range sortedCategoryKeys(request.Assessment.CategoryScores) {
			fmt.Fprintf(b, "| %s | %d |\n", category, request.Assessment.CategoryScores[category])
		}
		b.WriteString("\n")
	}
}

func renderAnalyzerResults(b *strings.Builder, request RenderRequest) {
	b.WriteString("## Analyzer Results\n\n")
	if len(request.AnalyzerResults) == 0 {
		b.WriteString("No analyzers ran.\n\n")
		return
	}
	b.WriteString("| Analyzer | Name | Version | Status | Findings |\n| --- | --- | --- | --- | --- |\n")
	for _, result := range request.AnalyzerResults {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %d |\n",
			markdownCell(result.AnalyzerID),
			markdownCell(result.Name),
			markdownCell(result.Version),
			markdownCell(string(result.Status)),
			len(result.Findings),
		)
	}
	b.WriteString("\n")
}

func renderFindings(b *strings.Builder, values []findings.Finding) {
	b.WriteString("## Findings\n\n")
	if len(values) == 0 {
		b.WriteString("No findings detected.\n\n")
		return
	}
	for _, finding := range values {
		fmt.Fprintf(b, "### %s: %s\n\n", strings.Title(string(finding.Severity)), markdownCell(finding.Title))
		fmt.Fprintf(b, "- Rule: `%s`\n", markdownCell(finding.RuleID))
		fmt.Fprintf(b, "- Analyzer: `%s`\n", markdownCell(finding.AnalyzerID))
		fmt.Fprintf(b, "- Category: `%s`\n", markdownCell(string(finding.Category)))
		fmt.Fprintf(b, "- Message: %s\n\n", markdownCell(finding.Message))
		b.WriteString("Evidence:\n\n")
		for _, evidence := range finding.Evidence {
			fmt.Fprintf(b, "- `%s`: %s", evidence.Type, markdownCell(evidence.Message))
			if evidence.Path != "" {
				fmt.Fprintf(b, " (`%s`)", markdownCell(evidence.Path))
			}
			if evidence.Value != "" {
				fmt.Fprintf(b, " - `%s`", markdownCell(evidence.Value))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
}

func renderRecommendations(b *strings.Builder, values []findings.Recommendation) {
	b.WriteString("## Recommendations\n\n")
	if len(values) == 0 {
		b.WriteString("No recommendations.\n\n")
		return
	}
	for _, recommendation := range values {
		fmt.Fprintf(b, "### %s\n\n", markdownCell(recommendation.Title))
		fmt.Fprintf(b, "- Finding: `%s`\n", markdownCell(recommendation.FindingID))
		fmt.Fprintf(b, "- Priority: `%s`\n", markdownCell(string(recommendation.Priority)))
		fmt.Fprintf(b, "- Action: %s\n", markdownCell(recommendation.Action))
		fmt.Fprintf(b, "- Rationale: %s\n\n", markdownCell(recommendation.Rationale))
	}
}

func renderMetadata(b *strings.Builder, format string, generatedAt time.Time) {
	b.WriteString("## Metadata\n\n")
	b.WriteString("| Field | Value |\n| --- | --- |\n")
	fmt.Fprintf(b, "| Format | %s |\n", markdownCell(format))
	fmt.Fprintf(b, "| Schema Version | %s |\n", schemaVersion)
	fmt.Fprintf(b, "| Generated At | %s |\n", generatedAt.Format(time.RFC3339))
	fmt.Fprintf(b, "| RepoCompass Version | %s |\n", "unknown")
}
