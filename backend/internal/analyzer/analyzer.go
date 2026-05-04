// Package analyzer defines the contracts used by repository analyzers.
package analyzer

import (
	"context"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Analyzer inspects a repository snapshot and returns structured analyzer output.
type Analyzer interface {
	Metadata() AnalyzerMetadata
	Analyze(ctx context.Context, input Input) (AnalyzerResult, error)
}

// Input contains read-only scan context passed to an analyzer.
type Input struct {
	Repository             repository.Repository
	Snapshot               snapshot.RepositorySnapshot
	EffectiveConfiguration config.EffectiveConfiguration
	RuleSet                rules.RuleSet
}

// AnalyzerMetadata identifies an analyzer in reports, logs, and registries.
type AnalyzerMetadata struct {
	ID      string
	Name    string
	Version string
}

// AnalyzerStatus describes the outcome of one analyzer execution.
type AnalyzerStatus string

const (
	AnalyzerStatusSuccess AnalyzerStatus = "success"
	AnalyzerStatusSkipped AnalyzerStatus = "skipped"
	AnalyzerStatusFailed  AnalyzerStatus = "failed"
)

// AnalyzerResult contains the output produced by one analyzer execution.
type AnalyzerResult struct {
	AnalyzerID   string
	Name         string
	Version      string
	Status       AnalyzerStatus
	Duration     time.Duration
	Findings     []findings.Finding
	Metadata     map[string]string
	ErrorMessage string
}
