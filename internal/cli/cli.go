package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/render"
)

// Run executes the gocensus command line interface.
func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, version string) int {
	if len(args) == 0 {
		args = []string{"scan", "."}
	} else if !isCommand(args[0]) {
		args = append([]string{"scan"}, args...)
	}

	switch args[0] {
	case "scan":
		return runScan(ctx, args[1:], stdout, stderr)
	case "packages", "files", "tests", "report":
		_, _ = fmt.Fprintf(stderr, "%s command is planned but not implemented yet\n", args[0])
		return 2
	case "version":
		_, _ = fmt.Fprintf(stdout, "gocensus %s\n", version)
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runScan(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	scanArgs, err := parseScanArgs(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 2
	}

	result, err := gocensus.Analyze(ctx, gocensus.Options{Root: scanArgs.root})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 1
	}

	if err := render.Result(stdout, result, scanArgs.format); err != nil {
		_, _ = fmt.Fprintf(stderr, "render failed: %v\n", err)
		return 1
	}
	return 0
}

type scanArgs struct {
	root   string
	format string
}

func parseScanArgs(args []string) (scanArgs, error) {
	parsed := scanArgs{
		root:   ".",
		format: "table",
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--format":
			i++
			if i >= len(args) {
				return scanArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			parsed.format = args[i]
		case strings.HasPrefix(arg, "--format="):
			parsed.format = strings.TrimPrefix(arg, "--format=")
		case strings.HasPrefix(arg, "-"):
			return scanArgs{}, fmt.Errorf("unknown flag %q", arg)
		default:
			if parsed.root != "." {
				return scanArgs{}, fmt.Errorf("unexpected argument %q", arg)
			}
			parsed.root = arg
		}
	}
	return parsed, nil
}

func isCommand(arg string) bool {
	switch arg {
	case "scan", "packages", "files", "tests", "report", "version", "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: gocensus <command> [path] [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  scan      Repository-level overview")
	_, _ = fmt.Fprintln(w, "  packages  Package-by-package metrics")
	_, _ = fmt.Fprintln(w, "  files     File-by-file metrics")
	_, _ = fmt.Fprintln(w, "  tests     Test/benchmark/example inventory")
	_, _ = fmt.Fprintln(w, "  report    Generate a Markdown report")
	_, _ = fmt.Fprintln(w, "  version   Print version/build info")
}
