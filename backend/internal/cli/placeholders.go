package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
	return newPlaceholderCmd(
		"scan",
		"Scan a repository with the RepoCompass pipeline.",
	)
}

func newReportCmd() *cobra.Command {
	return newPlaceholderCmd(
		"report",
		"Generate a RepoCompass report from collected data.",
	)
}

func newDoctorCmd() *cobra.Command {
	return newPlaceholderCmd(
		"doctor",
		"Run environment and setup diagnostics for RepoCompass.",
	)
}

func newPlaceholderCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "%s is not implemented yet.\n", cmd.Name())
			return nil
		},
	}
}
