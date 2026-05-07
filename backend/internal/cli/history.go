package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/history"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
	"github.com/spf13/cobra"
)

// newHistoryCmd creates the history command for persisted repository scans.
func newHistoryCmd() *cobra.Command {
	var format string
	var limit int

	cmd := &cobra.Command{
		Use:   "history <repository-id>",
		Short: "Show persisted scan history for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, cleanup, err := openHistoryStore(cmd.Context())
			if err != nil {
				return err
			}
			defer cleanup()

			items, err := store.ListScanHistory(cmd.Context(), args[0], limit)
			if err != nil {
				return fmt.Errorf("history failed: %w", err)
			}
			switch format {
			case "", "text":
				if len(items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No scans found for repository: %s\n", args[0])
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Scan History for repository: %s\n", args[0])
				for _, item := range items {
					fmt.Fprintf(cmd.OutOrStdout(), "- %s status=%s score=%d findings=%d completed=%s\n",
						item.ScanID,
						item.Status,
						item.Score,
						item.FindingCount,
						formatOptionalTime(item.CompletedAt),
					)
				}
				return nil
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(items)
			default:
				return fmt.Errorf("unsupported history format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format (text, json)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of scans to return")
	return cmd
}

// newFindingsCmd creates the findings command for persisted scan details.
func newFindingsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "findings <scan-id>",
		Short: "Show persisted finding details for a scan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, cleanup, err := openHistoryStore(cmd.Context())
			if err != nil {
				return err
			}
			defer cleanup()

			items, err := store.ListFindings(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("findings failed: %w", err)
			}
			switch format {
			case "", "text":
				if len(items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No findings found for scan: %s\n", args[0])
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Findings for scan: %s\n", args[0])
				for _, item := range items {
					fmt.Fprintf(cmd.OutOrStdout(), "- [%s] %s (%s)\n", item.Severity, item.Title, item.RuleID)
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", item.Message)
				}
				return nil
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(items)
			default:
				return fmt.Errorf("unsupported findings format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "", "Output format (text, json)")
	return cmd
}

// openHistoryStore opens the PostgreSQL store required by history commands.
func openHistoryStore(ctx context.Context) (history.Reader, func(), error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, func() {}, fmt.Errorf("DATABASE_URL is required")
	}
	db, err := pgstore.Open(dbURL)
	if err != nil {
		return nil, func() {}, fmt.Errorf("database connection failed: %w", err)
	}
	return pgstore.New(db), func() { _ = db.Close() }, nil
}

// formatOptionalTime formats nullable timestamps for text output.
func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
