package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version represents the current release version of RepoCompass.
// Bump this variable before cutting a new release.
var Version = "v0.1.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of RepoCompass",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("RepoCompass %s\n", Version)
		},
	}
}
