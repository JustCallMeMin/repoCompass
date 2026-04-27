package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
	"github.com/JustCallMeMin/repoCompass/backend/internal/storage/noop"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	var persist bool

	cmd := &cobra.Command{
		Use:   "scan <path>",
		Short: "Scan a repository locally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

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
			runner := scan.NewLocalScanRunner(provider, creator, resolver, logger, store)

			req := scan.RunRequest{
				Source: repository.RepositorySource{
					Type: repository.SourceTypeLocal,
					Path: path,
				},
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Starting scan for repository at: %s\n", path)

			result, err := runner.Run(cmd.Context(), req)
			if err != nil {
				var rcErr *rcerr.Error
				if errors.As(err, &rcErr) {
					return fmt.Errorf("scan failed [%s]: %s", rcErr.Code, rcErr.Message)
				}
				return fmt.Errorf("scan failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\nScan Summary:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Scan ID:      %s\n", result.Scan.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Snapshot ID:  %s\n", result.Scan.SnapshotID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:       %s\n", result.Scan.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Analyzers:    %d\n", result.Summary.AnalyzersProcessed)
			fmt.Fprintf(cmd.OutOrStdout(), "  Max File Size: %d bytes\n", result.EffectiveConfig.MaxFileSizeBytes)

			return nil
		},
	}

	cmd.Flags().BoolVar(&persist, "persist", false, "Persist scan results to the database (requires DATABASE_URL)")
	return cmd
}

// newDiscardLogger returns a no-op logger, useful in tests that only care about
// stdout/stderr output and not log events.
func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
}
