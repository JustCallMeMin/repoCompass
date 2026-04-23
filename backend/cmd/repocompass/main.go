package main

import (
	"fmt"
	"os"
)

const helpText = `repocompass

RepoCompass is an onboarding repository tool for new contributors.

Usage:
  repocompass [flags]

Flags:
  -h, --help    Show help for repocompass

Planned commands:
  scan          Planned for T0-005
  report        Planned for T0-005
  doctor        Planned for T0-005
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stdout, helpText)
		return 0
	}

	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			_, _ = fmt.Fprint(stdout, helpText)
			return 0
		default:
			_, _ = fmt.Fprintf(stderr, "unknown command: %s\n\n%s", arg, helpText)
			return 1
		}
	}

	return 0
}
