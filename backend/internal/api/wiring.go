package api

import (
	"log/slog"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/ci"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/contributing"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/readme"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/scripts"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

// NewPersistedLocalRunner builds the default persisted scan runner for API use.
func NewPersistedLocalRunner(store scan.Store, logger *slog.Logger) (*scan.LocalScanRunner, error) {
	registry, err := analyzer.NewAnalyzerRegistry(readme.New(), contributing.New(), ci.New(), scripts.New())
	if err != nil {
		return nil, err
	}
	return scan.NewLocalScanRunner(
		repository.NewLocalRepositoryProvider(),
		snapshot.NewCreator(),
		config.NewLocalConfigurationResolver(),
		logger,
		store,
		registry,
	), nil
}
