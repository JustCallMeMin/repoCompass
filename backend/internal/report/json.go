package report

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

// JSONRenderer renders machine-readable JSON reports.
type JSONRenderer struct {
	now func() time.Time
}

// NewJSONRenderer creates a JSON renderer.
func NewJSONRenderer() Renderer {
	return JSONRenderer{now: time.Now}
}

// Format returns the JSON report format ID.
func (r JSONRenderer) Format() Format {
	return FormatJSON
}

// Render converts scan result data into a stable JSON report.
func (r JSONRenderer) Render(ctx context.Context, request RenderRequest) (RenderResult, error) {
	if err := ctx.Err(); err != nil {
		return RenderResult{}, err
	}
	now := r.now
	if now == nil {
		now = time.Now
	}
	payload := buildJSONPayload(request, now().UTC())
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return RenderResult{}, err
	}
	content = append(content, '\n')
	return RenderResult{
		Format:      FormatJSON,
		Content:     content,
		ContentType: "application/json",
	}, nil
}

type jsonPayload struct {
	SchemaVersion   string               `json:"schema_version"`
	Repository      jsonRepository       `json:"repository"`
	Scan            jsonScan             `json:"scan"`
	Assessment      jsonAssessment       `json:"assessment"`
	Analyzers       []jsonAnalyzer       `json:"analyzers"`
	Findings        []jsonFinding        `json:"findings"`
	Recommendations []jsonRecommendation `json:"recommendations"`
	Metadata        jsonMetadata         `json:"metadata"`
}

type jsonRepository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	URL           string `json:"url,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

type jsonScan struct {
	ID          string `json:"id"`
	SnapshotID  string `json:"snapshot_id"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type jsonAssessment struct {
	OverallScore   int            `json:"overall_score"`
	Label          string         `json:"label"`
	FindingCount   int            `json:"finding_count"`
	SeverityCounts map[string]int `json:"severity_counts"`
	CategoryScores map[string]int `json:"category_scores"`
}

type jsonAnalyzer struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	DurationMS    int64  `json:"duration_ms"`
	FindingsCount int    `json:"findings_count"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

type jsonFinding struct {
	ID         string         `json:"id"`
	RuleID     string         `json:"rule_id"`
	AnalyzerID string         `json:"analyzer_id"`
	Severity   string         `json:"severity"`
	Title      string         `json:"title"`
	Message    string         `json:"message"`
	Category   string         `json:"category"`
	Status     string         `json:"status"`
	Evidence   []jsonEvidence `json:"evidence"`
}

type jsonEvidence struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Value   string `json:"value,omitempty"`
}

type jsonRecommendation struct {
	FindingID string `json:"finding_id"`
	Title     string `json:"title"`
	Action    string `json:"action"`
	Rationale string `json:"rationale"`
	Priority  string `json:"priority"`
}

type jsonMetadata struct {
	Format             string `json:"format"`
	GeneratedAt        string `json:"generated_at"`
	RepoCompassVersion string `json:"repocompass_version"`
}

func buildJSONPayload(request RenderRequest, generatedAt time.Time) jsonPayload {
	findingsList := collectFindings(request.AnalyzerResults)
	return jsonPayload{
		SchemaVersion: schemaVersion,
		Repository: jsonRepository{
			ID:            request.Repository.ID,
			Name:          request.Repository.Name,
			Provider:      string(request.Repository.Provider),
			URL:           request.Repository.URL,
			DefaultBranch: request.Repository.DefaultBranch,
		},
		Scan: jsonScan{
			ID:          request.Scan.ID,
			SnapshotID:  request.Scan.SnapshotID,
			Status:      string(request.Scan.Status),
			StartedAt:   formatTimePointer(request.Scan.StartTime),
			CompletedAt: formatTimePointer(request.Scan.EndTime),
		},
		Assessment:      buildJSONAssessment(request.Assessment),
		Analyzers:       buildJSONAnalyzers(request.AnalyzerResults),
		Findings:        buildJSONFindings(findingsList),
		Recommendations: buildJSONRecommendations(collectRecommendations(findingsList)),
		Metadata: jsonMetadata{
			Format:             string(FormatJSON),
			GeneratedAt:        generatedAt.Format(time.RFC3339),
			RepoCompassVersion: "unknown",
		},
	}
}

func buildJSONAssessment(value assessment.Assessment) jsonAssessment {
	return jsonAssessment{
		OverallScore:   value.OverallScore,
		Label:          string(value.Label),
		FindingCount:   value.FindingCount,
		SeverityCounts: severityCountsToJSON(value.SeverityCounts),
		CategoryScores: categoryScoresToJSON(value.CategoryScores),
	}
}

func buildJSONAnalyzers(values []analyzer.AnalyzerResult) []jsonAnalyzer {
	out := make([]jsonAnalyzer, 0, len(values))
	for _, value := range values {
		out = append(out, jsonAnalyzer{
			ID:            value.AnalyzerID,
			Name:          value.Name,
			Version:       value.Version,
			Status:        string(value.Status),
			DurationMS:    value.Duration.Milliseconds(),
			FindingsCount: len(value.Findings),
			ErrorMessage:  value.ErrorMessage,
		})
	}
	return out
}

func buildJSONFindings(values []findings.Finding) []jsonFinding {
	out := make([]jsonFinding, 0, len(values))
	for _, value := range values {
		out = append(out, jsonFinding{
			ID:         value.ID,
			RuleID:     value.RuleID,
			AnalyzerID: value.AnalyzerID,
			Severity:   string(value.Severity),
			Title:      value.Title,
			Message:    value.Message,
			Category:   string(value.Category),
			Status:     string(value.Status),
			Evidence:   buildJSONEvidence(value.Evidence),
		})
	}
	return out
}

func buildJSONEvidence(values []findings.Evidence) []jsonEvidence {
	out := make([]jsonEvidence, 0, len(values))
	for _, value := range values {
		out = append(out, jsonEvidence{
			Type:    string(value.Type),
			Message: value.Message,
			Path:    value.Path,
			Value:   value.Value,
		})
	}
	return out
}

func buildJSONRecommendations(values []findings.Recommendation) []jsonRecommendation {
	out := make([]jsonRecommendation, 0, len(values))
	for _, value := range values {
		out = append(out, jsonRecommendation{
			FindingID: value.FindingID,
			Title:     value.Title,
			Action:    value.Action,
			Rationale: value.Rationale,
			Priority:  string(value.Priority),
		})
	}
	return out
}

func severityCountsToJSON(values map[rules.Severity]int) map[string]int {
	out := make(map[string]int, len(values))
	for _, key := range sortedSeverityKeys(values) {
		out[string(key)] = values[key]
	}
	return out
}

func categoryScoresToJSON(values map[rules.Category]int) map[string]int {
	out := make(map[string]int, len(values))
	for _, key := range sortedCategoryKeys(values) {
		out[string(key)] = values[key]
	}
	return out
}
