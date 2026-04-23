package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repocompass",
		Short: "RepoCompass is an onboarding repository tool for new contributors.",
		Long:  "RepoCompass is an onboarding repository tool for new contributors.",
	}

	cmd.AddCommand(
		newScanCmd(),
		newReportCmd(),
		newDoctorCmd(),
	)

	return cmd
}
