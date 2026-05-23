package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/alecthomas/kong"
	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/render"
)

// Run executes the gocensus command line interface.
func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer, version string) (code int) {
	commandLine := commandLine{}
	parser, err := kong.New(&commandLine,
		kong.Name("gocensus"),
		kong.Description("Analyze Go repositories: files, lines, tests, ratios, and reports."),
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
	version string
}

type commandLine struct {
	analysisFlags `embed:""`

	Scan     scanCmd     `cmd:"" default:"withargs" help:"Repository-level overview."`
	Report   reportCmd   `cmd:"" help:"Generate a repository report."`
	Packages packagesCmd `cmd:"" help:"Package-by-package metrics."`
	Files    filesCmd    `cmd:"" help:"File-by-file metrics."`
	Tests    testsCmd    `cmd:"" help:"Test/benchmark/example inventory."`
	Version  versionCmd  `cmd:"" help:"Print version/build info."`
}

type analysisFlags struct {
	NoGitignore      bool     `name:"no-gitignore" help:"Do not read .gitignore exclude rules."`
	ExtraExcludes    []string `name:"exclude" placeholder:"PATTERN" help:"Exclude paths matching pattern; can be repeated."`
	IncludeGenerated bool     `name:"include-generated" help:"Include generated files in production totals."`
	IncludeMocks     bool     `name:"include-mocks" help:"Include mock files in production totals."`
}

type rootArg struct {
	Root string `arg:"" optional:"" default:"." type:"path" help:"Repository root to analyze."`
}

type scanCmd struct {
	rootArg
	Format string `enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type reportCmd struct {
	rootArg
	Format string `enum:"table,json,markdown" default:"markdown" help:"Output format: table, json, or markdown."`
	Output string `placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type packagesCmd struct {
	rootArg
	Sort string `enum:"path,test-ratio,prod-lines,test-lines" default:"path" help:"Sort packages by path, test-ratio, prod-lines, or test-lines."`
}

type filesCmd struct {
	rootArg
	Top int `default:"20" help:"Maximum number of files to print; use 0 for all files."`
}

type testsCmd struct {
	rootArg
}

type versionCmd struct{}

func (cmd *scanCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt.ctx, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output)
}

func (cmd *reportCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt.ctx, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output)
}

func (cmd *packagesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt.ctx, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderPackages(rt.stdout, result, cmd.Sort)
}

func (cmd *filesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt.ctx, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderFiles(rt.stdout, result, cmd.Top)
}

func (cmd *testsCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt.ctx, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderTests(rt.stdout, result)
}

func (cmd *versionCmd) Run(rt *runtime) error {
	_, err := fmt.Fprintf(rt.stdout, "gocensus %s\n", rt.version)
	return err
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

func analyze(ctx context.Context, root string, flags analysisFlags) (gocensus.Result, error) {
	return gocensus.Analyze(ctx, gocensus.Options{
		Root:             root,
		NoGitignore:      flags.NoGitignore,
		ExtraExcludes:    flags.ExtraExcludes,
		IncludeGenerated: flags.IncludeGenerated,
		IncludeMocks:     flags.IncludeMocks,
	})
}

func renderResult(w io.Writer, result gocensus.Result, format string, output string) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	if err := render.Result(writer, result, format); err != nil {
		return fmt.Errorf("render result: %w", err)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderPackages(w io.Writer, result gocensus.Result, sortBy string) error {
	if _, err := fmt.Fprintln(w, "Packages"); err != nil {
		return err
	}
	packages := sortedPackages(result.Packages, sortBy)
	for _, pkg := range packages {
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

func sortedPackages(packages []gocensus.PackageMetric, sortBy string) []gocensus.PackageMetric {
	sorted := slices.Clone(packages)
	slices.SortFunc(sorted, func(a gocensus.PackageMetric, b gocensus.PackageMetric) int {
		switch sortBy {
		case "test-ratio":
			if a.Ratios.TestToProductionEffective > b.Ratios.TestToProductionEffective {
				return -1
			}
			if a.Ratios.TestToProductionEffective < b.Ratios.TestToProductionEffective {
				return 1
			}
		case "prod-lines":
			if a.Lines.Production.Effective > b.Lines.Production.Effective {
				return -1
			}
			if a.Lines.Production.Effective < b.Lines.Production.Effective {
				return 1
			}
		case "test-lines":
			if a.Lines.Tests.Effective > b.Lines.Tests.Effective {
				return -1
			}
			if a.Lines.Tests.Effective < b.Lines.Tests.Effective {
				return 1
			}
		}
		if a.ImportPath < b.ImportPath {
			return -1
		}
		if a.ImportPath > b.ImportPath {
			return 1
		}
		return 0
	})
	return sorted
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
	knownTestCases := result.Tests.Tests + result.Tests.StaticSubtests
	knownBenchmarkCases := result.Tests.Benchmarks + result.Tests.StaticSubbenchmarks

	_, err := fmt.Fprintf(w, `Tests
  Known Test Cases         %d
  Top-level Tests          %d
  Static Subtests          %d
  Dynamic Subtest Sites    %d

Benchmarks
  Known Benchmark Cases    %d
  Top-level Benchmarks     %d
  Static Subbenchmarks     %d
  Dynamic Benchmark Sites  %d

Examples
  Examples                 %d
`,
		knownTestCases,
		result.Tests.Tests,
		result.Tests.StaticSubtests,
		result.Tests.DynamicSubtestSites,
		knownBenchmarkCases,
		result.Tests.Benchmarks,
		result.Tests.StaticSubbenchmarks,
		result.Tests.DynamicSubbenchmarkSites,
		result.Tests.Examples,
	)
	return err
}
