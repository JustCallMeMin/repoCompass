package scan

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// LocalScanRunner orchestrates the local scan execution lifecycle.
type LocalScanRunner struct {
	provider repository.LocalRepositoryProvider
	creator  snapshot.Creator
	resolver config.Resolver
	logger   *slog.Logger
	store    Store
	registry *analyzer.AnalyzerRegistry
}

// NewLocalScanRunner creates a new LocalScanRunner.
// logger is required; pass slog.Default() if no specific logger is needed.
// store is required; pass noop.New() if persistence is disabled.
func NewLocalScanRunner(
	provider repository.LocalRepositoryProvider,
	creator snapshot.Creator,
	resolver config.Resolver,
	logger *slog.Logger,
	store Store,
	registry *analyzer.AnalyzerRegistry,
) *LocalScanRunner {
	if logger == nil {
		logger = slog.Default()
	}
	if registry == nil {
		registry, _ = analyzer.NewAnalyzerRegistry()
	}
	return &LocalScanRunner{
		provider: provider,
		creator:  creator,
		resolver: resolver,
		logger:   logger,
		store:    store,
		registry: registry,
	}
}

// Run executes a scan locally by resolving the repository, creating a snapshot, and returning a summary.
func (r *LocalScanRunner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	s := Scan{
		ID:     generateScanID(),
		Status: StatusCreated,
	}

	// Bind scan_id to every log call in this run.
	l := r.logger.With(
		slog.String("scan_id", s.ID),
		slog.String("repo_path", request.Source.Path),
	)

	// Transition to running
	if err := s.TransitionTo(StatusRunning); err != nil {
		return RunResult{Scan: s}, rcerr.New(rcerr.CodeInternalError, "failed to start scan", err)
	}

	start := time.Now()
	s.StartTime = &start

	l.InfoContext(ctx, "scan started",
		slog.String("operation", "scan_start"),
	)

	// 0. Resolve Configuration
	effConfig, err := r.resolver.ResolveConfig(ctx, request.Source.Path, request.ConfigOverrides)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := wrapIfNeeded(err, rcerr.CodeConfigResolveFailed, "failed to resolve configuration")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	l.InfoContext(ctx, "config resolved",
		slog.String("operation", "config_resolve"),
		slog.Int64("max_file_size_bytes", effConfig.MaxFileSizeBytes),
	)

	ruleSet, err := rules.ResolveRuleSet(rules.DefaultRuleSet(), effConfig)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to resolve ruleset", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s, EffectiveConfig: effConfig}, wrapped
	}

	// 1. Resolve Repository
	resolution, err := r.provider.Resolve(ctx, request.Source)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := wrapIfNeeded(err, rcerr.CodeRepoResolveFailed, "failed to resolve repository")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	l.InfoContext(ctx, "repository resolved",
		slog.String("operation", "repo_resolve"),
		slog.String("repository_id", resolution.Repository.ID),
		slog.String("repository_name", resolution.Repository.Name),
	)

	// 2. Create Snapshot
	snap, err := r.creator.Create(ctx, snapshot.CreateRequest{
		Repository:       resolution.Repository,
		SourceType:       snapshot.SourceTypeLocal,
		SnapshotMetadata: resolution.SnapshotMetadata,
	})
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := wrapIfNeeded(err, rcerr.CodeSnapshotCreateFailed, "failed to create snapshot")
		code, _ := rcerr.CodeOf(wrapped)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("error_id", string(code)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	s.SnapshotID = snap.ID

	l.InfoContext(ctx, "snapshot created",
		slog.String("operation", "snapshot_create"),
		slog.String("snapshot_id", snap.ID),
		slog.String("repository_id", resolution.Repository.ID),
	)

	// 3. Execute analyzers.
	analyzerResults, err := r.runAnalyzers(ctx, analyzer.Input{
		Repository:             resolution.Repository,
		Snapshot:               snap,
		EffectiveConfiguration: effConfig,
		RuleSet:                ruleSet,
	}, effConfig, l)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to resolve analyzers", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s, AnalyzerResults: analyzerResults}, wrapped
	}

	allFindings := collectFindings(analyzerResults)

	// Fetch active policy for the org. Ignore errors as policy might be missing or persistence disabled.
	var policy assessment.OrgPolicy
	if r.store != nil {
		if p, err := r.store.GetActiveAssessmentPolicy(ctx, request.Source.OrganizationID); err == nil {
			policy = p
		}
	}

	scanAssessment, err := assessment.NewEngine().Assess(allFindings, policy)
	if err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to assess scan findings", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s, AnalyzerResults: analyzerResults}, wrapped
	}

	// 4. Complete Scan.
	if err := s.TransitionTo(StatusCompleted); err != nil {
		s.ErrorDetails = err.Error()
		_ = s.TransitionTo(StatusFailed)
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to complete scan", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return RunResult{Scan: s}, wrapped
	}

	endTime := time.Now()
	s.EndTime = &endTime

	result := RunResult{
		Scan:       s,
		Repository: resolution.Repository,
		Snapshot:   snap,
		Summary: Summary{
			AnalyzersProcessed: len(analyzerResults),
			FindingCount:       len(allFindings),
			AssessmentScore:    scanAssessment.OverallScore,
		},
		EffectiveConfig: effConfig,
		AnalyzerResults: analyzerResults,
		Assessment:      scanAssessment,
	}

	if err := r.store.SaveRunResult(ctx, result); err != nil {
		s.ErrorDetails = err.Error()
		s.Status = StatusFailed
		result.Scan = s
		wrapped := rcerr.New(rcerr.CodeScanExecutionFailed, "failed to persist scan result", err)
		l.ErrorContext(ctx, "scan failed",
			slog.String("operation", "scan_failed"),
			slog.String("repository_id", resolution.Repository.ID),
			slog.String("snapshot_id", snap.ID),
			slog.String("error_id", string(rcerr.CodeScanExecutionFailed)),
			slog.String("error_msg", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return result, wrapped
	}

	l.InfoContext(ctx, "scan completed",
		slog.String("operation", "scan_completed"),
		slog.String("repository_id", resolution.Repository.ID),
		slog.String("snapshot_id", snap.ID),
		slog.String("status", string(StatusCompleted)),
		slog.Int64("duration_ms", endTime.Sub(start).Milliseconds()),
	)

	return result, nil
}

// runAnalyzers executes selected analyzers sequentially and isolates analyzer-level failures.
func (r *LocalScanRunner) runAnalyzers(
	ctx context.Context,
	input analyzer.Input,
	effConfig config.EffectiveConfiguration,
	logger *slog.Logger,
) ([]analyzer.AnalyzerResult, error) {
	selected, err := r.registry.Resolve(effConfig)
	if err != nil {
		return nil, err
	}

	results := make([]analyzer.AnalyzerResult, 0, len(selected))
	for _, current := range selected {
		metadata := current.Metadata()
		start := time.Now()
		result, err := current.Analyze(ctx, input)
		if err != nil {
			result = analyzer.AnalyzerResult{
				AnalyzerID:   metadata.ID,
				Name:         metadata.Name,
				Version:      metadata.Version,
				Status:       analyzer.AnalyzerStatusFailed,
				Duration:     time.Since(start),
				ErrorMessage: err.Error(),
			}
			logger.ErrorContext(ctx, "analyzer failed",
				slog.String("operation", "analyzer_failed"),
				slog.String("analyzer_id", metadata.ID),
				slog.String("error_msg", err.Error()),
			)
		} else {
			if result.AnalyzerID == "" {
				result.AnalyzerID = metadata.ID
			}
			if result.Name == "" {
				result.Name = metadata.Name
			}
			if result.Version == "" {
				result.Version = metadata.Version
			}
			if result.Status == "" {
				result.Status = analyzer.AnalyzerStatusSuccess
			}
			if result.Duration == 0 {
				result.Duration = time.Since(start)
			}
			logger.InfoContext(ctx, "analyzer completed",
				slog.String("operation", "analyzer_completed"),
				slog.String("analyzer_id", result.AnalyzerID),
				slog.String("status", string(result.Status)),
				slog.Int("finding_count", len(result.Findings)),
				slog.Int64("duration_ms", result.Duration.Milliseconds()),
			)
		}
		results = append(results, result)
	}
	return results, nil
}

func collectFindings(results []analyzer.AnalyzerResult) []findings.Finding {
	var collected []findings.Finding
	for _, result := range results {
		collected = append(collected, result.Findings...)
	}
	return collected
}

func generateScanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "scan_" + hex.EncodeToString(b)
}

// wrapIfNeeded wraps err with a new rcerr.Error only if err is not already an *rcerr.Error.
// This prevents double-wrapping when sub-packages already return structured errors.
func wrapIfNeeded(err error, code rcerr.ErrorCode, message string) error {
	var rcErr *rcerr.Error
	if errors.As(err, &rcErr) {
		return err
	}
	return rcerr.New(code, message, err)
}
