// Package ci provides the built-in CI workflow analyzer.
package ci

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
	analyzerID      = "ci"
	analyzerName    = "CI Workflow Analyzer"
	analyzerVersion = "0.1.0"
	ruleWorkflow    = "ci.workflow.exists"
	ruleTestJob     = "ci.test_job.exists"
)

var testSignals = []string{"go test", "npm test", "pytest", "cargo test", "mvn test", "gradle test", "test:"}

// Analyzer checks GitHub Actions workflow presence and basic test command signals.
type Analyzer struct{}

// New creates a CI workflow analyzer.
func New() Analyzer {
	return Analyzer{}
}

// Metadata returns stable analyzer identity.
func (Analyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{ID: analyzerID, Name: analyzerName, Version: analyzerVersion}
}

// Analyze checks CI workflow files without parsing YAML fully.
func (Analyzer) Analyze(ctx context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	start := time.Now()
	if err := ctx.Err(); err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	activeRules := activeRuleSet(input.RuleSet, ruleWorkflow, ruleTestJob)
	if len(activeRules) == 0 {
		return analyzer.AnalyzerResult{
			AnalyzerID: analyzerID,
			Name:       analyzerName,
			Version:    analyzerVersion,
			Status:     analyzer.AnalyzerStatusSkipped,
			Duration:   time.Since(start),
			Metadata:   map[string]string{"reason": "rules_disabled"},
		}, nil
	}

	repoPath := input.Repository.LocalPath
	if repoPath == "" {
		repoPath = input.Snapshot.SnapshotMetadata["local_path"]
	}
	if repoPath == "" {
		return analyzer.AnalyzerResult{}, fmt.Errorf("ci analyzer requires local repository path")
	}

	files, err := workflowFiles(repoPath)
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
			"workflow_count": fmt.Sprintf("%d", len(files)),
		},
	}

	if len(files) == 0 {
		if activeRules[ruleWorkflow] {
			finding, err := buildMissingWorkflowFinding()
			if err != nil {
				return analyzer.AnalyzerResult{}, err
			}
			result.Findings = append(result.Findings, finding)
		}
		return result, nil
	}

	hasTestSignal, err := workflowsContainTestSignal(files)
	if err != nil {
		return analyzer.AnalyzerResult{}, err
	}
	if activeRules[ruleTestJob] && !hasTestSignal {
		finding, err := buildMissingTestJobFinding()
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}
	result.Duration = time.Since(start)
	return result, nil
}

func workflowFiles(repoPath string) ([]string, error) {
	info, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("stat repository path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repository path must be a directory")
	}

	workflowDir := filepath.Join(repoPath, ".github", "workflows")
	info, err = os.Stat(workflowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat workflow directory: %w", err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return nil, fmt.Errorf("read workflow directory: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			files = append(files, filepath.Join(workflowDir, name))
		}
	}
	return files, nil
}

func workflowsContainTestSignal(files []string) (bool, error) {
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return false, fmt.Errorf("read workflow file %s: %w", filepath.Base(file), err)
		}
		text := strings.ToLower(string(content))
		for _, signal := range testSignals {
			if strings.Contains(text, signal) {
				return true, nil
			}
		}
	}
	return false, nil
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

func buildMissingWorkflowFinding() (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleWorkflow, ".github/workflows")
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleWorkflow,
		Target:     ".github/workflows",
		Severity:   rules.SeverityMedium,
		Category:   rules.CategoryCI,
		Title:      "CI workflow is missing",
		Message:    "The repository does not include a GitHub Actions workflow file.",
		Evidence: []findings.Evidence{
			findings.FileMissingEvidence(".github/workflows", "No .yml or .yaml workflow file was found under .github/workflows."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add CI workflow",
				"Create a GitHub Actions workflow under .github/workflows that runs repository tests.",
				"Contributors need automated validation before changes are merged.",
				findings.RecommendationPriorityMedium,
			),
		},
	})
}

func buildMissingTestJobFinding() (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleTestJob, ".github/workflows")
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleTestJob,
		Target:     ".github/workflows",
		Severity:   rules.SeverityMedium,
		Category:   rules.CategoryCI,
		Title:      "CI test job is missing",
		Message:    "Existing workflow files do not include a recognizable test command.",
		Evidence: []findings.Evidence{
			findings.PatternMissingEvidence(".github/workflows", strings.Join(testSignals, ", "), "No supported test command signal was found in workflow files."),
		},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add test job to CI",
				"Add an explicit test command such as go test ./... or npm test to a workflow job.",
				"CI should verify contributor changes automatically.",
				findings.RecommendationPriorityMedium,
			),
		},
	})
}
