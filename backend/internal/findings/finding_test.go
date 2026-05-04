package findings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestFindingValidateAcceptsValidFinding(t *testing.T) {
	finding := validFinding()

	if err := finding.Validate(); err != nil {
		t.Fatalf("expected valid finding, got error: %v", err)
	}
}

func TestEvidenceValidateAcceptsSupportedTypes(t *testing.T) {
	tests := []findings.Evidence{
		findings.NewEvidence(findings.EvidenceTypeFileExists, "README.md exists.", "README.md", ""),
		findings.NewEvidence(findings.EvidenceTypeFileMissing, "README.md is missing.", "README.md", ""),
		findings.NewEvidence(findings.EvidenceTypePatternMatch, "Setup section exists.", "README.md", "## Setup"),
		findings.NewEvidence(findings.EvidenceTypePatternMissing, "Setup section is missing.", "README.md", "## Setup"),
		findings.NewEvidence(findings.EvidenceTypeMetadata, "Default branch is main.", "", "default_branch=main"),
	}

	for _, evidence := range tests {
		t.Run(string(evidence.Type), func(t *testing.T) {
			if err := evidence.Validate(); err != nil {
				t.Fatalf("expected valid evidence, got error: %v", err)
			}
		})
	}
}

func TestEvidenceValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		evidence findings.Evidence
	}{
		{
			name:     "missing type",
			evidence: findings.NewEvidence("", "README.md is missing.", "README.md", ""),
		},
		{
			name:     "missing message",
			evidence: findings.NewEvidence(findings.EvidenceTypeFileMissing, "", "README.md", ""),
		},
		{
			name:     "file evidence missing path",
			evidence: findings.NewEvidence(findings.EvidenceTypeFileMissing, "README.md is missing.", "", ""),
		},
		{
			name:     "pattern evidence missing path",
			evidence: findings.NewEvidence(findings.EvidenceTypePatternMissing, "Setup section is missing.", "", "## Setup"),
		},
		{
			name:     "pattern evidence missing value",
			evidence: findings.NewEvidence(findings.EvidenceTypePatternMissing, "Setup section is missing.", "README.md", ""),
		},
		{
			name:     "metadata evidence missing value",
			evidence: findings.NewEvidence(findings.EvidenceTypeMetadata, "Default branch metadata exists.", "", ""),
		},
		{
			name:     "absolute path",
			evidence: findings.NewEvidence(findings.EvidenceTypeFileMissing, "README.md is missing.", filepath.Join(os.TempDir(), "repo", "README.md"), ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.evidence.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRecommendationValidateAcceptsValidRecommendation(t *testing.T) {
	recommendation := validRecommendation()

	if err := recommendation.Validate(); err != nil {
		t.Fatalf("expected valid recommendation, got error: %v", err)
	}
}

func TestRecommendationValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*findings.Recommendation)
	}{
		{name: "missing finding id", mutate: func(r *findings.Recommendation) { r.FindingID = "" }},
		{name: "missing title", mutate: func(r *findings.Recommendation) { r.Title = "" }},
		{name: "missing action", mutate: func(r *findings.Recommendation) { r.Action = "" }},
		{name: "missing rationale", mutate: func(r *findings.Recommendation) { r.Rationale = "" }},
		{name: "invalid priority", mutate: func(r *findings.Recommendation) { r.Priority = findings.RecommendationPriority("urgent") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := validRecommendation()
			tt.mutate(&recommendation)
			if err := recommendation.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNewFindingSetsOpenStatus(t *testing.T) {
	finding := findings.NewFinding(
		"finding_readme_exists",
		"readme.exists",
		"readme",
		rules.SeverityHigh,
		rules.CategoryDocumentation,
		"README file is missing",
		"Repository does not include a root README file.",
		[]findings.Evidence{validEvidence()},
	)

	if finding.Status != findings.FindingStatusOpen {
		t.Fatalf("expected status %q, got %q", findings.FindingStatusOpen, finding.Status)
	}
}

func TestFindingValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*findings.Finding)
	}{
		{name: "missing id", mutate: func(f *findings.Finding) { f.ID = "" }},
		{name: "missing rule id", mutate: func(f *findings.Finding) { f.RuleID = "" }},
		{name: "missing analyzer id", mutate: func(f *findings.Finding) { f.AnalyzerID = "" }},
		{name: "missing title", mutate: func(f *findings.Finding) { f.Title = "" }},
		{name: "missing message", mutate: func(f *findings.Finding) { f.Message = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finding := validFinding()
			tt.mutate(&finding)
			if err := finding.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestFindingValidateRejectsMissingOrInvalidEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*findings.Finding)
	}{
		{name: "missing evidence", mutate: func(f *findings.Finding) { f.Evidence = nil }},
		{name: "invalid evidence", mutate: func(f *findings.Finding) { f.Evidence = []findings.Evidence{{Type: findings.EvidenceTypeFileMissing}} }},
		{name: "invalid recommendation", mutate: func(f *findings.Finding) { f.Recommendations = []findings.Recommendation{{FindingID: f.ID}} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finding := validFinding()
			tt.mutate(&finding)
			if err := finding.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestFindingValidateRejectsInvalidEnums(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*findings.Finding)
	}{
		{name: "invalid severity", mutate: func(f *findings.Finding) { f.Severity = rules.Severity("critical") }},
		{name: "invalid category", mutate: func(f *findings.Finding) { f.Category = rules.Category("security") }},
		{name: "invalid status", mutate: func(f *findings.Finding) { f.Status = findings.FindingStatus("closed") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finding := validFinding()
			tt.mutate(&finding)
			if err := finding.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func validFinding() findings.Finding {
	return findings.Finding{
		ID:              "finding_readme_exists",
		RuleID:          "readme.exists",
		AnalyzerID:      "readme",
		Severity:        rules.SeverityHigh,
		Title:           "README file is missing",
		Message:         "Repository does not include a root README file.",
		Category:        rules.CategoryDocumentation,
		Status:          findings.FindingStatusOpen,
		Evidence:        []findings.Evidence{validEvidence()},
		Recommendations: []findings.Recommendation{validRecommendation()},
	}
}

func validEvidence() findings.Evidence {
	return findings.NewEvidence(
		findings.EvidenceTypeFileMissing,
		"README.md was not found at the repository root.",
		"README.md",
		"",
	)
}

func validRecommendation() findings.Recommendation {
	return findings.NewRecommendation(
		"finding_readme_exists",
		"Add a root README",
		"Create README.md with project purpose, setup steps, and test commands.",
		"New contributors need one stable entry point before changing code.",
		findings.RecommendationPriorityHigh,
	)
}
