package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
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
		return runResult(ctx, args[1:], stdout, stderr, "table")
	case "report":
		return runResult(ctx, args[1:], stdout, stderr, "markdown")
	case "packages":
		return runView(ctx, args[1:], stdout, stderr, "packages")
	case "files":
		return runView(ctx, args[1:], stdout, stderr, "files")
	case "tests":
		return runView(ctx, args[1:], stdout, stderr, "tests")
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

func runResult(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, defaultFormat string) int {
	parsed, err := parseCommonArgs(args, defaultFormat)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 2
	}

	result, err := analyze(ctx, parsed)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 1
	}

	var out bytes.Buffer
	writer := io.Writer(stdout)
	if parsed.output != "" {
		writer = &out
	}
	if err := render.Result(writer, result, parsed.format); err != nil {
		_, _ = fmt.Fprintf(stderr, "render failed: %v\n", err)
		return 1
	}
	if parsed.output != "" {
		if err := os.WriteFile(parsed.output, out.Bytes(), 0o644); err != nil {
			_, _ = fmt.Fprintf(stderr, "write failed: %v\n", err)
			return 1
		}
	}
	return 0
}

func runView(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, view string) int {
	parsed, err := parseCommonArgs(args, "table")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 2
	}

	result, err := analyze(ctx, parsed)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "command failed: %v\n", err)
		return 1
	}

	switch view {
	case "packages":
		err = renderPackages(stdout, result)
	case "files":
		err = renderFiles(stdout, result, parsed.top)
	case "tests":
		err = renderTests(stdout, result)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "render failed: %v\n", err)
		return 1
	}
	return 0
}

type commonArgs struct {
	root             string
	format           string
	output           string
	top              int
	sort             string
	useGitignore     bool
	extraExcludes    []string
	includeGenerated bool
	includeMocks     bool
}

func parseCommonArgs(args []string, defaultFormat string) (commonArgs, error) {
	parsed := commonArgs{
		root:         ".",
		format:       defaultFormat,
		top:          20,
		sort:         "path",
		useGitignore: true,
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--format":
			i++
			if i >= len(args) {
				return commonArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			parsed.format = args[i]
		case strings.HasPrefix(arg, "--format="):
			parsed.format = strings.TrimPrefix(arg, "--format=")
		case arg == "--output":
			i++
			if i >= len(args) {
				return commonArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			parsed.output = args[i]
		case strings.HasPrefix(arg, "--output="):
			parsed.output = strings.TrimPrefix(arg, "--output=")
		case arg == "--top":
			i++
			if i >= len(args) {
				return commonArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			top, err := strconv.Atoi(args[i])
			if err != nil {
				return commonArgs{}, fmt.Errorf("invalid --top value: %w", err)
			}
			parsed.top = top
		case strings.HasPrefix(arg, "--top="):
			top, err := strconv.Atoi(strings.TrimPrefix(arg, "--top="))
			if err != nil {
				return commonArgs{}, fmt.Errorf("invalid --top value: %w", err)
			}
			parsed.top = top
		case arg == "--sort":
			i++
			if i >= len(args) {
				return commonArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			parsed.sort = args[i]
		case strings.HasPrefix(arg, "--sort="):
			parsed.sort = strings.TrimPrefix(arg, "--sort=")
		case arg == "--no-gitignore":
			parsed.useGitignore = false
		case arg == "--exclude":
			i++
			if i >= len(args) {
				return commonArgs{}, fmt.Errorf("%s requires a value", arg)
			}
			parsed.extraExcludes = append(parsed.extraExcludes, args[i])
		case strings.HasPrefix(arg, "--exclude="):
			parsed.extraExcludes = append(parsed.extraExcludes, strings.TrimPrefix(arg, "--exclude="))
		case arg == "--include-generated":
			parsed.includeGenerated = true
		case arg == "--include-mocks":
			parsed.includeMocks = true
		case strings.HasPrefix(arg, "-"):
			return commonArgs{}, fmt.Errorf("unknown flag %q", arg)
		default:
			if parsed.root != "." {
				return commonArgs{}, fmt.Errorf("unexpected argument %q", arg)
			}
			parsed.root = arg
		}
	}
	return parsed, nil
}

func analyze(ctx context.Context, parsed commonArgs) (gocensus.Result, error) {
	return gocensus.Analyze(ctx, gocensus.Options{
		Root:             parsed.root,
		NoGitignore:      !parsed.useGitignore,
		ExtraExcludes:    parsed.extraExcludes,
		IncludeGenerated: parsed.includeGenerated,
		IncludeMocks:     parsed.includeMocks,
	})
}

func renderPackages(w io.Writer, result gocensus.Result) error {
	if _, err := fmt.Fprintln(w, "Packages"); err != nil {
		return err
	}
	for _, pkg := range result.Packages {
		if _, err := fmt.Fprintf(w, "  %s  prod=%d  test=%d  ratio=%.2f:1\n",
			pkg.ImportPath,
			pkg.Lines.Production.Effective,
			pkg.Lines.Tests.Effective,
			pkg.Ratios.TestToProductionEffective,
		); err != nil {
			return err
		}
	}
	return nil
}

func renderFiles(w io.Writer, result gocensus.Result, top int) error {
	if _, err := fmt.Fprintln(w, "Files"); err != nil {
		return err
	}
	for i, file := range result.FileMetrics {
		if top > 0 && i >= top {
			break
		}
		if _, err := fmt.Fprintf(w, "  %s  %s  raw=%d  effective=%d\n",
			file.Path, file.Kind, file.RawLines, file.CodeLines); err != nil {
			return err
		}
	}
	return nil
}

func renderTests(w io.Writer, result gocensus.Result) error {
	_, err := fmt.Fprintf(w, "Tests\n  Test funcs  %d\n  Benchmarks  %d\n  Examples    %d\n",
		result.Tests.Tests,
		result.Tests.Benchmarks,
		result.Tests.Examples,
	)
	return err
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
