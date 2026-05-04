// Package report contains report generation building blocks and output shaping.
package report

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// Format identifies a supported report output format.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
)

// RenderRequest contains the complete scan result context needed by renderers.
type RenderRequest struct {
	Scan            scan.Scan
	Repository      repository.Repository
	Snapshot        snapshot.RepositorySnapshot
	AnalyzerResults []analyzer.AnalyzerResult
	Assessment      assessment.Assessment
	EffectiveConfig config.EffectiveConfiguration
}

// RenderResult contains rendered report bytes and output metadata.
type RenderResult struct {
	Format      Format
	Content     []byte
	ContentType string
}

// Renderer renders a scan result into one report format.
type Renderer interface {
	Format() Format
	Render(ctx context.Context, request RenderRequest) (RenderResult, error)
}
