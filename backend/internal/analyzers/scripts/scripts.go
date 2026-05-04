// Package scripts provides the built-in developer scripts analyzer.
package scripts

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
	analyzerID      = "scripts"
	analyzerName    = "Scripts Analyzer"
	analyzerVersion = "0.1.0"
	ruleDev         = "scripts.dev.exists"
	ruleTest        = "scripts.test_command.exists"
	ruleFmt         = "scripts.fmt_command.exists"
	ruleLint        = "scripts.lint_command.exists"
)

var testSignals = []string{"test", "go test", "npm test", "pytest", "cargo test", "mvn test", "gradle test"}
var fmtSignals = []string{"fmt", "format", "gofmt", "prettier", "black", "rustfmt"}
var lintSignals = []string{"lint", "golangci-lint", "eslint", "ruff", "clippy"}

// Analyzer checks for contributor-facing developer commands.
type Analyzer struct{}

// New creates a scripts analyzer.
func New() Analyzer {
	return Analyzer{}
}

// Metadata returns stable analyzer identity.
func (Analyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{ID: analyzerID, Name: analyzerName, Version: analyzerVersion}
}

// Analyze checks whether a repository exposes basic dev/test/fmt/lint commands.
func (Analyzer) Analyze(ctx context.Context, input analyzer.Input) (analyzer.AnalyzerResult, error) {
	start := time.Now()
	if err := ctx.Err(); err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	activeRules := activeRuleSet(input.RuleSet, ruleDev, ruleTest, ruleFmt, ruleLint)
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
		return analyzer.AnalyzerResult{}, fmt.Errorf("scripts analyzer requires local repository path")
	}

	files, err := candidateFiles(repoPath)
	if err != nil {
		return analyzer.AnalyzerResult{}, err
	}

	result := analyzer.AnalyzerResult{
		AnalyzerID: analyzerID,
		Name:       analyzerName,
		Version:    analyzerVersion,
		Status:     analyzer.AnalyzerStatusSuccess,
		Duration:   time.Since(start),
		Metadata:   map[string]string{"script_source_count": fmt.Sprintf("%d", len(files))},
	}

	if len(files) == 0 {
		if activeRules[ruleDev] {
			finding, err := buildFinding(ruleDev, "developer commands", rules.SeverityLow, "Developer command entrypoint is missing", "The repository does not expose a Makefile, package.json, or scripts directory.", findings.FileMissingEvidence("Makefile/package.json/scripts", "No supported developer command entrypoint was found."))
			if err != nil {
				return analyzer.AnalyzerResult{}, err
			}
			result.Findings = append(result.Findings, finding)
		}
		return result, nil
	}

	text, err := readCandidates(files)
	if err != nil {
		return analyzer.AnalyzerResult{}, err
	}
	if activeRules[ruleTest] && !containsAny(text, testSignals) {
		finding, err := commandFinding(ruleTest, "test", testSignals, rules.SeverityMedium, "Test command is missing", "The repository command entrypoints do not include a recognizable test command.")
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}
	if activeRules[ruleFmt] && !containsAny(text, fmtSignals) {
		finding, err := commandFinding(ruleFmt, "fmt", fmtSignals, rules.SeverityLow, "Format command is missing", "The repository command entrypoints do not include a recognizable format command.")
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}
	if activeRules[ruleLint] && !containsAny(text, lintSignals) {
		finding, err := commandFinding(ruleLint, "lint", lintSignals, rules.SeverityLow, "Lint command is missing", "The repository command entrypoints do not include a recognizable lint command.")
		if err != nil {
			return analyzer.AnalyzerResult{}, err
		}
		result.Findings = append(result.Findings, finding)
	}
	result.Duration = time.Since(start)
	return result, nil
}

func candidateFiles(repoPath string) ([]string, error) {
	info, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("stat repository path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repository path must be a directory")
	}

	var files []string
	for _, name := range []string{"Makefile", "package.json"} {
		path := filepath.Join(repoPath, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			files = append(files, path)
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat command file %s: %w", name, err)
		}
	}

	scriptsDir := filepath.Join(repoPath, "scripts")
	info, err = os.Stat(scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, fmt.Errorf("stat scripts directory: %w", err)
	}
	if !info.IsDir() {
		return files, nil
	}
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("read scripts directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(scriptsDir, entry.Name()))
		}
	}
	return files, nil
}

func readCandidates(files []string) (string, error) {
	var b strings.Builder
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read command file %s: %w", filepath.Base(file), err)
		}
		b.WriteString(strings.ToLower(string(content)))
		b.WriteByte('\n')
	}
	return b.String(), nil
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

func commandFinding(ruleID string, target string, signals []string, severity rules.Severity, title string, message string) (findings.Finding, error) {
	return buildFinding(ruleID, target, severity, title, message, findings.PatternMissingEvidence("Makefile/package.json/scripts", strings.Join(signals, ", "), "No supported command signal was found."))
}

func buildFinding(ruleID string, target string, severity rules.Severity, title string, message string, evidence findings.Evidence) (findings.Finding, error) {
	findingID := findings.BuildFindingID(analyzerID, ruleID, target)
	return findings.BuildFinding(findings.FindingSpec{
		AnalyzerID: analyzerID,
		RuleID:     ruleID,
		Target:     target,
		Severity:   severity,
		Category:   rules.CategoryWorkflow,
		Title:      title,
		Message:    message,
		Evidence:   []findings.Evidence{evidence},
		Recommendations: []findings.Recommendation{
			findings.RecommendationForFinding(
				findingID,
				"Add "+target+" command",
				"Expose a "+target+" command through Makefile, package.json, or scripts/.",
				"Contributors need stable commands for local repository checks.",
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
