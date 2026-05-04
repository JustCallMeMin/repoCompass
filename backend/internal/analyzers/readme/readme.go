// Package readme provides the built-in README analyzer.
package readme

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

const (
	analyzerID      = "readme"
	analyzerName    = "README Analyzer"
	analyzerVersion = "0.1.0"
	readmeExistsID  = "readme.exists"
)

var readmeFilenames = []string{"README.md", "README.markdown", "README.txt", "README"}

// Analyzer checks whether a repository has a root README file.
type Analyzer struct{}

// New creates a README analyzer.
func New() Analyzer {
	return Analyzer{}
}

// Metadata returns stable analyzer identity.
func (Analyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{
		ID:      analyzerID,
		Name:    analyzerName,
		Version: analyzerVersion,
	}
}

// Analyze checks README existence without mutating repository contents.
func (Analyzer) Analyze(ctx context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	start := time.Now()
	if err := ctx.Err(); err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	if !ruleEnabled(input.RuleSet, readmeExistsID) {
		return analyzer.AnalyzerResult{
			AnalyzerID: analyzerID,
			Name:       analyzerName,
			Version:    analyzerVersion,
			Status:     analyzer.AnalyzerStatusSkipped,
			Duration:   time.Since(start),
			Metadata: map[string]string{
				"reason":  "rule_disabled",
				"rule_id": readmeExistsID,
			},
		}, nil
	}

	repoPath := input.Repository.LocalPath
	if repoPath == "" {
		repoPath = input.Snapshot.SnapshotMetadata["local_path"]
	}
	if repoPath == "" {
		return analyzer.AnalyzerResult{}, fmt.Errorf("readme analyzer requires local repository path")
	}

	present, err := readmePresent(repoPath)
	if err != nil {
		return analyzer.AnalyzerResult{}, err
	}
	if present {
		return analyzer.AnalyzerResult{
			AnalyzerID: analyzerID,
			Name:       analyzerName,
			Version:    analyzerVersion,
			Status:     analyzer.AnalyzerStatusSuccess,
			Duration:   time.Since(start),
			Metadata: map[string]string{
				"readme_present": "true",
			},
		}, nil
	}

	findingID := findings.BuildFindingID(analyzerID, readmeExistsID, "README")
	finding, err := findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     readmeExistsID,
		Target:     "README",
		Severity:   rules.SeverityHigh,
		Category:   rules.CategoryDocumentation,
		Title:      "README file is missing",
		Message:    "The repository does not contain a README file at its root.",
		Evidence: []findings.Evidence{
			findings.FileMissingEvidence("README", "No accepted README filename was found at the repository root."),
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
		return analyzer.AnalyzerResult{}, err
	}

	return analyzer.AnalyzerResult{
		AnalyzerID: analyzerID,
		Name:       analyzerName,
		Version:    analyzerVersion,
		Status:     analyzer.AnalyzerStatusSuccess,
		Duration:   time.Since(start),
		Findings:   []findings.Finding{finding},
		Metadata: map[string]string{
			"readme_present": "false",
		},
	}, nil
}

func ruleEnabled(ruleSet rules.RuleSet, ruleID string) bool {
	for _, rule := range ruleSet.Rules {
		if rule.ID == ruleID {
			return true
		}
	}
	return false
}

func readmePresent(repoPath string) (bool, error) {
	info, err := os.Stat(repoPath)
	if err != nil {
		return false, fmt.Errorf("stat repository path: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("repository path must be a directory")
	}

	for _, filename := range readmeFilenames {
		path := filepath.Join(repoPath, filename)
		info, err := os.Stat(path)
		if err == nil {
			return !info.IsDir(), nil
		}
		if os.IsNotExist(err) {
			continue
		}
		return false, fmt.Errorf("stat README candidate %s: %w", filename, err)
	}
	return false, nil
}
