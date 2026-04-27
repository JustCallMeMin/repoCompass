package cli

import (
	"bytes"
	"testing"
)

func TestRootCommandHelpIncludesExpectedSubcommands(t *testing.T) {
	cmd := NewRootCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected help command to succeed: %v", err)
	}

	output := stdout.String()
	assertContains(t, output, "repocompass")
	assertContains(t, output, "scan")
	assertContains(t, output, "report")
	assertContains(t, output, "doctor")
}

func TestPlaceholderCommandsReturnExpectedMessages(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "report placeholder",
			args:     []string{"report"},
			expected: "report is not implemented yet.\n",
		},
		{
			name:     "doctor placeholder",
			args:     []string{"doctor"},
			expected: "doctor is not implemented yet.\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs(tc.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected placeholder command to succeed: %v", err)
			}

			if got := stdout.String(); got != tc.expected {
				t.Fatalf("unexpected stdout: got %q want %q", got, tc.expected)
			}

			if gotErr := stderr.String(); gotErr != "" {
				t.Fatalf("expected empty stderr, got %q", gotErr)
			}
		})
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !bytes.Contains([]byte(haystack), []byte(needle)) {
		t.Fatalf("expected %q to contain %q", haystack, needle)
	}
}
