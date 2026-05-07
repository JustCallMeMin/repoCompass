package assessment_test

import (
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestEngineAssessNoFindingsGivesExcellentScore(t *testing.T) {
	result, err := assessment.NewEngine().Assess(nil, assessment.OrgPolicy{})
	if err != nil {
		t.Fatalf("expected assessment to succeed: %v", err)
	}

	if result.OverallScore != 100 {
		t.Fatalf("expected score 100, got %d", result.OverallScore)
	}
	if result.Label != assessment.ScoreLabelExcellent {
		t.Fatalf("expected label %q, got %q", assessment.ScoreLabelExcellent, result.Label)
	}
	if result.FindingCount != 0 {
		t.Fatalf("expected 0 findings, got %d", result.FindingCount)
	}
}

func TestEngineAssessAppliesSeverityPenalties(t *testing.T) {
	result, err := assessment.NewEngine().Assess([]findings.Finding{
		validFinding("high", rules.SeverityHigh, rules.CategoryDocumentation),
		validFinding("medium", rules.SeverityMedium, rules.CategoryCI),
		validFinding("low", rules.SeverityLow, rules.CategoryWorkflow),
	}, assessment.OrgPolicy{})
	if err != nil {
		t.Fatalf("expected assessment to succeed: %v", err)
	}

	if result.OverallScore != 60 {
		t.Fatalf("expected score 60, got %d", result.OverallScore)
	}
	if result.Label != assessment.ScoreLabelFair {
		t.Fatalf("expected label %q, got %q", assessment.ScoreLabelFair, result.Label)
	}
	if result.SeverityCounts[rules.SeverityHigh] != 1 {
		t.Fatalf("expected one high finding, got %d", result.SeverityCounts[rules.SeverityHigh])
	}
	if result.SeverityCounts[rules.SeverityMedium] != 1 {
		t.Fatalf("expected one medium finding, got %d", result.SeverityCounts[rules.SeverityMedium])
	}
	if result.SeverityCounts[rules.SeverityLow] != 1 {
		t.Fatalf("expected one low finding, got %d", result.SeverityCounts[rules.SeverityLow])
	}
}

func TestEngineAssessFloorsScoreAtZero(t *testing.T) {
	values := make([]findings.Finding, 0, 5)
	for i := 0; i < 5; i++ {
		values = append(values, validFinding(string(rune('a'+i)), rules.SeverityHigh, rules.CategoryDocumentation))
	}

	result, err := assessment.NewEngine().Assess(values, assessment.OrgPolicy{})
	if err != nil {
		t.Fatalf("expected assessment to succeed: %v", err)
	}
	if result.OverallScore != 0 {
		t.Fatalf("expected score floor 0, got %d", result.OverallScore)
	}
	if result.Label != assessment.ScoreLabelPoor {
		t.Fatalf("expected label %q, got %q", assessment.ScoreLabelPoor, result.Label)
	}
}

func TestEngineAssessBuildsCategoryBreakdown(t *testing.T) {
	result, err := assessment.NewEngine().Assess([]findings.Finding{
		validFinding("doc", rules.SeverityHigh, rules.CategoryDocumentation),
		validFinding("ci", rules.SeverityMedium, rules.CategoryCI),
		validFinding("workflow", rules.SeverityLow, rules.CategoryWorkflow),
	}, assessment.OrgPolicy{})
	if err != nil {
		t.Fatalf("expected assessment to succeed: %v", err)
	}

	if result.CategoryScores[rules.CategoryDocumentation] != 75 {
		t.Fatalf("expected documentation score 75, got %d", result.CategoryScores[rules.CategoryDocumentation])
	}
	if result.CategoryScores[rules.CategoryCI] != 90 {
		t.Fatalf("expected CI score 90, got %d", result.CategoryScores[rules.CategoryCI])
	}
	if result.CategoryScores[rules.CategoryWorkflow] != 95 {
		t.Fatalf("expected workflow score 95, got %d", result.CategoryScores[rules.CategoryWorkflow])
	}
	if result.CategoryBreakdown[rules.CategoryDocumentation].SeverityCounts[rules.SeverityHigh] != 1 {
		t.Fatalf("expected documentation breakdown to count high severity")
	}
}

func TestEngineAssessAppliesOrgPolicy(t *testing.T) {
	readmeFinding := validFinding("readme", rules.SeverityLow, rules.CategoryDocumentation)
	readmeFinding.RuleID = "readme.exists"

	// Missing README but policy requires it -> extra -50 penalty.
	result, _ := assessment.NewEngine().Assess([]findings.Finding{readmeFinding}, assessment.OrgPolicy{
		RequireReadme: true,
	})
	if result.OverallScore != 45 { // 100 - 5 (low) - 50
		t.Fatalf("expected score 45 with RequireReadme, got %d", result.OverallScore)
	}

	// Score >= MinimumScore -> normal label.
	result2, _ := assessment.NewEngine().Assess(nil, assessment.OrgPolicy{
		MinimumScore: 90,
	})
	if result2.Label != assessment.ScoreLabelExcellent {
		t.Fatalf("expected label excellent, got %q", result2.Label)
	}

	// Score < MinimumScore -> poor label.
	result3, _ := assessment.NewEngine().Assess([]findings.Finding{
		validFinding("high", rules.SeverityHigh, rules.CategoryDocumentation), // Score 75
	}, assessment.OrgPolicy{
		MinimumScore: 80,
	})
	if result3.Label != assessment.ScoreLabelPoor {
		t.Fatalf("expected label poor, got %q", result3.Label)
	}
}

func TestEngineAssessRejectsInvalidFinding(t *testing.T) {
	_, err := assessment.NewEngine().Assess([]findings.Finding{{ID: "invalid"}}, assessment.OrgPolicy{})
	if err == nil {
		t.Fatal("expected invalid finding to fail")
	}
}

func validFinding(id string, severity rules.Severity, category rules.Category) findings.Finding {
	return findings.Finding{
		ID:         "finding_" + id,
		RuleID:     "rule." + id,
		AnalyzerID: "analyzer",
		Severity:   severity,
		Title:      "Finding " + id,
		Message:    "Finding message " + id,
		Category:   category,
		Status:     findings.FindingStatusOpen,
		Evidence: []findings.Evidence{
			findings.NewEvidence(findings.EvidenceTypeMetadata, "Metadata evidence.", "", "key=value"),
		},
	}
}
