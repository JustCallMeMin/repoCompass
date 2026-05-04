package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/ci"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/contributing"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/readme"
	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzers/scripts"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/report"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
	"github.com/JustCallMeMin/repoCompass/backend/internal/storage/noop"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	var persist bool
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "scan <path>",
		Short: "Scan a repository locally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			if output != "" && format == "" {
				return fmt.Errorf("--output requires --format")
			}

			var renderer report.Renderer
			if format != "" {
				registry, err := report.DefaultRegistry()
				if err != nil {
					return fmt.Errorf("scan failed: cannot initialize report renderer registry: %w", err)
				}
				renderer, err = registry.RendererFor(report.Format(format))
				if err != nil {
					return err
				}
			}

			// Structured logs go to stderr; scan results go to stdout.
			logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			// Initialise the store: postgres when --persist is set, noop otherwise.
			var store scan.Store
			if persist {
				dbURL := os.Getenv("DATABASE_URL")
				if dbURL == "" {
					return fmt.Errorf("--persist requires DATABASE_URL to be set")
				}
				db, err := pgstore.Open(dbURL)
				if err != nil {
					return fmt.Errorf("scan failed: cannot connect to database: %w", err)
				}
				defer db.Close()
				store = pgstore.New(db)
				logger.Info("persistence enabled", "store", "postgres")
			} else {
				store = noop.New()
			}

			provider := repository.NewLocalRepositoryProvider()
			creator := snapshot.NewCreator()
			resolver := config.NewLocalConfigurationResolver()
			registry, err := analyzer.NewAnalyzerRegistry(readme.New(), contributing.New(), ci.New(), scripts.New())
			if err != nil {
				return fmt.Errorf("scan failed: cannot initialize analyzer registry: %w", err)
			}
			runner := scan.NewLocalScanRunner(provider, creator, resolver, logger, store, registry)

			req := scan.RunRequest{
				Source: repository.RepositorySource{
					Type: repository.SourceTypeLocal,
					Path: path,
				},
			}

			if format == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Starting scan for repository at: %s\n", path)
			}

			result, err := runner.Run(cmd.Context(), req)
			if err != nil {
				var rcErr *rcerr.Error
				if errors.As(err, &rcErr) {
					return fmt.Errorf("scan failed [%s]: %s", rcErr.Code, rcErr.Message)
				}
				return fmt.Errorf("scan failed: %w", err)
			}

			if renderer != nil {
				rendered, err := renderer.Render(cmd.Context(), report.RenderRequest{
					Scan:            result.Scan,
					Repository:      result.Repository,
					Snapshot:        result.Snapshot,
					AnalyzerResults: result.AnalyzerResults,
					Assessment:      result.Assessment,
					EffectiveConfig: result.EffectiveConfig,
				})
				if err != nil {
					return fmt.Errorf("scan failed: render report: %w", err)
				}
				if output != "" {
					if err := os.WriteFile(output, rendered.Content, 0644); err != nil {
						return fmt.Errorf("scan failed: write report: %w", err)
					}
					return nil
				}
				_, err = cmd.OutOrStdout().Write(rendered.Content)
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\nScan Summary:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Scan ID:      %s\n", result.Scan.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Repository ID: %s\n", result.Repository.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Snapshot ID:  %s\n", result.Scan.SnapshotID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:       %s\n", result.Scan.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Analyzers:    %d\n", result.Summary.AnalyzersProcessed)
			fmt.Fprintf(cmd.OutOrStdout(), "  Findings:     %d\n", result.Summary.FindingCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Score:        %d/100 (%s)\n", result.Assessment.OverallScore, result.Assessment.Label)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max File Size: %d bytes\n", result.EffectiveConfig.MaxFileSizeBytes)

			return nil
		},
	}

	cmd.Flags().BoolVar(&persist, "persist", false, "Persist scan results to the database (requires DATABASE_URL)")
	cmd.Flags().StringVar(&format, "format", "", "Render full report in a supported format (markdown, json)")
	cmd.Flags().StringVar(&output, "output", "", "Write rendered report to a file (requires --format)")
	return cmd
}

// newDiscardLogger returns a no-op logger, useful in tests that only care about
// stdout/stderr output and not log events.
func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}
