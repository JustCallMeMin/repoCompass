// Package contributing provides the built-in CONTRIBUTING analyzer.
package contributing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

const (
	analyzerID      = "contributing"
	analyzerName    = "CONTRIBUTING Analyzer"
	analyzerVersion = "0.1.0"
	ruleExists      = "contributing.exists"
	ruleSetup       = "contributing.setup_instructions"
	ruleTest        = "contributing.test_instructions"
)

var contributingFilenames = []string{"CONTRIBUTING.md", "CONTRIBUTING.markdown", "CONTRIBUTING.txt", "CONTRIBUTING"}
var setupPatterns = []string{"setup", "install", "local development", "getting started"}
var testPatterns = []string{"test", "go test", "npm test", "pytest", "cargo test", "mvn test", "gradle test"}

// Analyzer checks contributor guidance files.
type Analyzer struct{}

// New creates a CONTRIBUTING analyzer.
func New() Analyzer {
	return Analyzer{}
}

// Metadata returns stable analyzer identity.
func (Analyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{ID: analyzerID, Name: analyzerName, Version: analyzerVersion}
}

// Analyze checks CONTRIBUTING file presence and minimal setup/test guidance.
func (Analyzer) Analyze(ctx context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	start := time.Now()
	if err := ctx.Err(); err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	activeRules := activeRuleSet(input.RuleSet, ruleExists, ruleSetup, ruleTest)
	if len(activeRules) == 0 {
		return skipped(start), nil
	}

	repoPath := input.Repository.LocalPath
	if repoPath == "" {
		repoPath = input.Snapshot.SnapshotMetadata["local_path"]
	}
	if repoPath == "" {
		return analyzer.AnalyzerResult{}, fmt.Errorf("contributing analyzer requires local repository path")
	}

	path, found, err := findContributingFile(repoPath)
	if err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	result := analyzer.AnalyzerResult{
		AnalyzerID: analyzerID,
		Name:       analyzerName,
		Version:    analyzerVersion,
		Status:     analyzer.AnalyzerStatusSuccess,
		Duration:   time.Since(start),
		Metadata: map[string]string{
			"contributing_present": boolString(found),
		},
	}

	if !found {
		if activeRules[ruleExists] {
			finding, err := buildMissingFileFinding()
			if err != nil {
				return analyzer.AnalyzerResult{}, err
			}
			result.Findings = append(result.Findings, finding)
		}
		return result, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return analyzer.AnalyzerResult{}, fmt.Errorf("read CONTRIBUTING file: %w", err)
	}
	text := strings.ToLower(string(content))
	relativePath := filepath.Base(path)

	if activeRules[ruleSetup] && !containsAny(text, setupPatterns) {
		finding, err := buildMissingSetupFinding(relativePath)
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}
	if activeRules[ruleTest] && !containsAny(text, testPatterns) {
		finding, err := buildMissingTestFinding(relativePath)
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}

	result.Duration = time.Since(start)
	return result, nil
}

func skipped(start time.Time) analyzer.AnalyzerResult {
	return analyzer.AnalyzerResult{
		AnalyzerID: analyzerID,
		Name:       analyzerName,
		Version:    analyzerVersion,
		Status:     analyzer.AnalyzerStatusSkipped,
		Duration:   time.Since(start),
		Metadata: map[string]string{
			"reason": "rules_disabled",
		},
	}
}

func activeRuleSet(ruleSet rules.RuleSet, ids ...string) map[string]bool {
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = false
	}
	for _, rule := range ruleSet.Rules {
		if _, ok := wanted[rule.ID]; ok {
			wanted[rule.ID] = true
		}
	}
	active := make(map[string]bool, len(ids))
	for id, ok := range wanted {
		if ok {
			active[id] = true
		}
	}
	return active
}

func findContributingFile(repoPath string) (string, bool, error) {
	info, err := os.Stat(repoPath)
	if err != nil {
		return "", false, fmt.Errorf("stat repository path: %w", err)
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf("repository path must be a directory")
	}
	for _, filename := range contributingFilenames {
		path := filepath.Join(repoPath, filename)
		info, err := os.Stat(path)
		if err == nil {
			return path, !info.IsDir(), nil
		}
		if os.IsNotExist(err) {
			continue
		}
		return "", false, fmt.Errorf("stat CONTRIBUTING candidate %s: %w", filename, err)
	}
	return "", false, nil
}

func buildMissingFileFinding() (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleExists, "CONTRIBUTING")
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleExists,
		Target:     "CONTRIBUTING",
		Severity:   rules.SeverityMedium,
		Category:   rules.CategoryDocumentation,
		Title:      "CONTRIBUTING file is missing",
		Message:    "The repository does not include contributor guidance at its root.",
		Evidence: []findings.Evidence{
			findings.FileMissingEvidence("CONTRIBUTING", "No accepted CONTRIBUTING filename was found at the repository root."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add contributor guidance",
				"Create CONTRIBUTING.md with setup, test, and contribution workflow instructions.",
				"New contributors need clear instructions before opening changes.",
				findings.RecommendationPriorityMedium,
			),
		},
	})
}

func buildMissingSetupFinding(path string) (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleSetup, path)
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleSetup,
		Target:     path,
		Severity:   rules.SeverityMedium,
		Category:   rules.CategoryDocumentation,
		Title:      "CONTRIBUTING setup guidance is missing",
		Message:    "The CONTRIBUTING file does not include basic setup or installation guidance.",
		Evidence: []findings.Evidence{
			findings.PatternMissingEvidence(path, strings.Join(setupPatterns, ", "), "No setup guidance pattern was found in the CONTRIBUTING file."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add setup instructions",
				"Add a setup or local development section to CONTRIBUTING.md.",
				"Contributors need to know how to prepare the repository locally.",
				findings.RecommendationPriorityMedium,
			),
		},
	})
}

func buildMissingTestFinding(path string) (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleTest, path)
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleTest,
		Target:     path,
		Severity:   rules.SeverityMedium,
		Category:   rules.CategoryDocumentation,
		Title:      "CONTRIBUTING test guidance is missing",
		Message:    "The CONTRIBUTING file does not include basic test command guidance.",
		Evidence: []findings.Evidence{
			findings.PatternMissingEvidence(path, strings.Join(testPatterns, ", "), "No test guidance pattern was found in the CONTRIBUTING file."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add test instructions",
				"Add the repository test command to CONTRIBUTING.md.",
				"Contributors need a clear validation command before submitting changes.",
				findings.RecommendationPriorityMedium,
			),
		},
	})
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
