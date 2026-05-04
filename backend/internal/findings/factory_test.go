package findings_test

import (
	"strings"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestBuildFindingIDIsStable(t *testing.T) {
	first := findings.BuildFindingID("readme", "readme.exists", "README.md")
	second := findings.BuildFindingID("readme", "readme.exists", "README.md")

	if first != second {
		t.Fatalf("expected stable finding ID, got %q and %q", first, second)
	}
	if !strings.HasPrefix(first, "finding_readme_readme_exists_") {
		t.Fatalf("unexpected finding ID prefix: %q", first)
	}
}

func TestBuildFindingIDChangesWithTarget(t *testing.T) {
	first := findings.BuildFindingID("readme", "readme.exists", "README.md")
	second := findings.BuildFindingID("readme", "readme.exists", "README.txt")

	if first == second {
		t.Fatalf("expected different target to produce different ID, got %q", first)
	}
}

func TestBuildFindingBuildsValidFinding(t *testing.T) {
	findingID := findings.BuildFindingID("readme", "readme.exists", "README.md")
	finding, err := findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: "readme",
		RuleID:     "readme.exists",
		Target:     "README.md",
		Severity:   rules.SeverityHigh,
		Category:   rules.CategoryDocumentation,
		Title:      "README file is missing",
		Message:    "The repository does not contain a README file at its root.",
		Evidence: []findings.Evidence{
			findings.FileMissingEvidence("README.md", "README.md was not found at the repository root."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add a root README",
				"Create README.md with project purpose, setup steps, and test commands.",
				"New contributors need one stable entry point before changing code.",
				findings.RecommendationPriorityHigh,
			),
		},
	})
	if err != nil {
		t.Fatalf("expected valid finding, got error: %v", err)
	}

	if finding.ID != findingID {
		t.Fatalf("expected finding ID %q, got %q", findingID, finding.ID)
	}
	if len(finding.Evidence) != 1 {
		t.Fatalf("expected 1 evidence item, got %d", len(finding.Evidence))
	}
	if len(finding.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(finding.Recommendations))
	}
}

func TestBuildFindingRejectsInvalidSpec(t *testing.T) {
	_, err := findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: "readme",
		RuleID:     "readme.exists",
		Target:     "README.md",
		Severity:   rules.SeverityHigh,
		Category:   rules.CategoryDocumentation,
		Title:      "README file is missing",
		Message:    "The repository does not contain a README file at its root.",
	})
	if err == nil {
		t.Fatal("expected missing evidence to fail")
	}
}

func TestBuildFindingRejectsInvalidRecommendation(t *testing.T) {
	_, err := findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: "readme",
		RuleID:     "readme.exists",
		Target:     "README.md",
		Severity:   rules.SeverityHigh,
		Category:   rules.CategoryDocumentation,
		Title:      "README file is missing",
		Message:    "The repository does not contain a README file at its root.",
		Evidence: []findings.Evidence{
			findings.FileMissingEvidence("README.md", "README.md was not found at the repository root."),
		},
		Recommendations: []findings.Recommendation{
			{FindingID: "finding_readme"},
		},
	})
	if err == nil {
		t.Fatal("expected invalid recommendation to fail")
	}
}
