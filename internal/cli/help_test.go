package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/cli"
)

func TestRunRootHelpIncludesCommandsAndCommonFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"--help"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"scan",
		"diff",
		"hotspots",
		"tests",
		"--no-gitignore",
		"-x, --exclude",
		"--include-generated",
		"--include-testdata",
		"--color",
		"--no-color",
		"production, test, generated, and mock code",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunSubcommandHelpIncludesCommandFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", "--help"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"scan",
		"--format",
		"--output",
		"--exclude",
		"--color",
		"--no-color",
		"-f, --format",
		"-o, --output",
		"known test cases",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunAllSubcommandHelpIsSelfExplaining(t *testing.T) {
	tests := []struct {
		command string
		wants   []string
	}{
		{
			command: "report",
			wants: []string{
				"saved in CI artifacts",
				"-f, --format",
				"-o, --output",
			},
		},
		{
			command: "packages",
			wants: []string{
				"effective production and test lines",
				"-s, --sort",
				"test-ratio",
			},
		},
		{
			command: "files",
			wants: []string{
				"classification and line counts",
				"-n, --top",
				"large files",
			},
		},
		{
			command: "tests",
			wants: []string{
				"top-level tests",
				"statically countable subtests",
				"dynamic subtest sites",
			},
		},
		{
			command: "diff",
			wants: []string{
				"Compare scan metrics",
				"--base",
				"--head",
				"--include-generated",
				"--include-mocks",
				"--include-testdata",
				"generated",
				"mock",
			},
		},
		{
			command: "hotspots",
			wants: []string{
				"human-authored production Go files",
				"--by",
				"--since",
				"--until",
				"--include-generated",
				"--include-mocks",
				"--include-testdata",
				"generated",
				"mock",
			},
		},
		{
			command: "who",
			wants: []string{
				"Git authors",
				"--go-only",
				"--include-generated",
				"--include-mocks",
				"--include-testdata",
				"--by",
				"--since",
				"--until",
				"scan scope",
				"Go paths",
				"active days",
			},
		},
		{
			command: "version",
			wants: []string{
				"Print the gocensus version",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := cli.Run(context.Background(), []string{tt.command, "--help"}, &stdout, &stderr, "dev")

			if code != 0 {
				t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
			}
			for _, want := range tt.wants {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want %q", stdout.String(), want)
				}
			}
		})
	}
}

func TestRunHelpCommandShowsSubcommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"help", "tests"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"tests",
		"Summarize test inventory",
		"--no-gitignore",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}
