// Package analyzer defines the contracts used by repository analyzers.
package analyzer

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Analyzer inspects a repository snapshot and returns structured analyzer output.
type Analyzer interface {
	Metadata() Metadata
	Analyze(ctx context.Context, input Input) (Result, error)
}

// Input contains read-only scan context passed to an analyzer.
type Input struct {
	Repository             repository.Repository
	Snapshot               snapshot.RepositorySnapshot
	EffectiveConfiguration config.EffectiveConfiguration
}

// Metadata identifies an analyzer in reports, logs, and registries.
type Metadata struct {
	ID      string
	Name    string
	Version string
}

// Result contains the output produced by one analyzer execution.
type Result struct {
	AnalyzerID string
	Metadata   map[string]string
	Findings   []Finding
}

// Finding is the minimal issue shape returned by analyzers in T2-004.
type Finding struct {
	RuleID  string
	Title   string
	Message string
}
