package scan

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// RunRequest contains the input data required to run a scan.
type RunRequest struct {
	Source          repository.RepositorySource
	ConfigOverrides config.Config
}

// Summary contains structured output from the scan execution.
type Summary struct {
	AnalyzersProcessed int
	FindingCount       int
	AssessmentScore    int
}

// RunResult contains the output data produced by a scan execution.
type RunResult struct {
	Scan            Scan
	Repository      repository.Repository
	Snapshot        snapshot.RepositorySnapshot
	Summary         Summary
	EffectiveConfig config.EffectiveConfiguration
	AnalyzerResults []analyzer.AnalyzerResult
	Assessment      assessment.Assessment
}

// ScanRunner defines the contract for executing repository scans.
type ScanRunner interface {
	Run(ctx context.Context, request RunRequest) (RunResult, error)
}
