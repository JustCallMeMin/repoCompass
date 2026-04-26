package scan

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

// RunRequest contains the input data required to run a scan.
type RunRequest struct {
	Source repository.RepositorySource
}

// Summary contains structured output from the scan execution.
type Summary struct {
	AnalyzersProcessed int
}

// RunResult contains the output data produced by a scan execution.
type RunResult struct {
	Scan    Scan
	Summary Summary
}

// ScanRunner defines the contract for executing repository scans.
type ScanRunner interface {
	Run(ctx context.Context, request RunRequest) (RunResult, error)
}
