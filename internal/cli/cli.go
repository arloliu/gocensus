package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/alecthomas/kong"
)

// Run executes the gocensus command line interface.
func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, version string) (code int) {
	commandLine := commandLine{}
	parser, err := kong.New(&commandLine,
		kong.Name("gocensus"),
		kong.Description("Analyze Go repositories and separate production, test, generated, and mock code into readable counts, ratios, and reports."),
		kong.Writers(stdout, stderr),
		kong.UsageOnError(),
		kong.Exit(func(exitCode int) {
			panic(kongExit(exitCode))
		}),
	)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 1
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		exit, ok := recovered.(kongExit)
		if !ok {
			panic(recovered)
		}
		code = int(exit)
	}()

	kongCtx, err := parser.Parse(normalizeHelpArgs(args))
	if err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
			return 2
		}
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 2
	}

	if err := kongCtx.Run(&runtime{
		ctx:     ctx,
		stdout:  stdout,
		environ: os.Environ(),
		version: version,
	}); err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 1
	}
	return 0
}

type kongExit int

type runtime struct {
	ctx     context.Context
	stdout  io.Writer
	environ []string
	version string
}

func normalizeHelpArgs(args []string) []string {
	if len(args) == 0 || args[0] != "help" {
		return args
	}
	if len(args) == 1 {
		return []string{"--help"}
	}
	normalized := slices.Clone(args[1:])
	normalized = append(normalized, "--help")
	return normalized
}
