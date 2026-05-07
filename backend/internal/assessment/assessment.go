// Package assessment contains assessment models and coordination logic.
package assessment

import (
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

const (
	maxScore      = 100
	highPenalty   = 25
	mediumPenalty = 10
	lowPenalty    = 5
)

// ScoreLabel describes the human-readable score band.
type ScoreLabel string

const (
	ScoreLabelExcellent ScoreLabel = "excellent"
	ScoreLabelGood      ScoreLabel = "good"
	ScoreLabelFair      ScoreLabel = "fair"
	ScoreLabelPoor      ScoreLabel = "poor"
)

// CategoryBreakdown contains score details for one finding category.
type CategoryBreakdown struct {
	Category       rules.Category
	Score          int
	FindingCount   int
	SeverityCounts map[rules.Severity]int
}

// Assessment summarizes repository onboarding health from findings.
type Assessment struct {
	OverallScore      int
	Label             ScoreLabel
	FindingCount      int
	SeverityCounts    map[rules.Severity]int
	CategoryScores    map[rules.Category]int
	CategoryBreakdown map[rules.Category]CategoryBreakdown
}

// Engine calculates deterministic assessment scores from findings.
type Engine struct{}

// NewEngine creates an assessment engine.
func NewEngine() Engine {
	return Engine{}
}

// OrgPolicy represents organization-level assessment constraints.
type OrgPolicy struct {
	MinimumScore  int  `json:"minimum_score,omitempty"`
	RequireReadme bool `json:"require_readme,omitempty"`
}

// Assess calculates an assessment from validated findings.
func (Engine) Assess(values []findings.Finding, policy OrgPolicy) (Assessment, error) {
	for _, finding := range values {
		if err := finding.Validate(); err != nil {
			return Assessment{}, fmt.Errorf("invalid finding %q: %w", finding.ID, err)
		}
	}

	severityCounts := make(map[rules.Severity]int)
	categoryFindings := make(map[rules.Category][]findings.Finding)
	for _, finding := range values {
		severityCounts[finding.Severity]++
		categoryFindings[finding.Category] = append(categoryFindings[finding.Category], finding)
	}

	overallScore := score(values)

	hasMissingReadme := false
	for _, finding := range values {
		if finding.RuleID == "readme.exists" {
			hasMissingReadme = true
			break
		}
	}

	if policy.RequireReadme && hasMissingReadme {
		overallScore -= 50
		if overallScore < 0 {
			overallScore = 0
		}
	}

	label := labelForScore(overallScore)
	if policy.MinimumScore > 0 && overallScore < policy.MinimumScore {
		label = ScoreLabelPoor
	}

	assessment := Assessment{
		OverallScore:      overallScore,
		Label:             label,
		FindingCount:      len(values),
		SeverityCounts:    severityCounts,
		CategoryScores:    make(map[rules.Category]int, len(categoryFindings)),
		CategoryBreakdown: make(map[rules.Category]CategoryBreakdown, len(categoryFindings)),
	}

	for category, findingsForCategory := range categoryFindings {
		categoryScore := score(findingsForCategory)
		categorySeverityCounts := make(map[rules.Severity]int)
		for _, finding := range findingsForCategory {
			categorySeverityCounts[finding.Severity]++
		}
		assessment.CategoryScores[category] = categoryScore
		assessment.CategoryBreakdown[category] = CategoryBreakdown{
			Category:       category,
			Score:          categoryScore,
			FindingCount:   len(findingsForCategory),
			SeverityCounts: categorySeverityCounts,
		}
	}

	return assessment, nil
}

func score(values []findings.Finding) int {
	score := maxScore
	for _, finding := range values {
		score -= penaltyForSeverity(finding.Severity)
	}
	if score < 0 {
		return 0
	}
	return score
}

func penaltyForSeverity(severity rules.Severity) int {
	switch severity {
	case rules.SeverityHigh:
		return highPenalty
	case rules.SeverityMedium:
		return mediumPenalty
	case rules.SeverityLow:
		return lowPenalty
	default:
		return 0
	}
}

func labelForScore(score int) ScoreLabel {
	switch {
	case score >= 90:
		return ScoreLabelExcellent
	case score >= 75:
		return ScoreLabelGood
	case score >= 50:
		return ScoreLabelFair
	default:
		return ScoreLabelPoor
	}
}
