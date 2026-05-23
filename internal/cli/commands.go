package cli

import (
	"fmt"

	"github.com/arloliu/gocensus"
)

type commandLine struct {
	analysisFlags `embed:""`

	Scan     scanCmd     `cmd:"" default:"withargs" help:"Print the main repository census."`
	Report   reportCmd   `cmd:"" help:"Write a repository report."`
	Packages packagesCmd `cmd:"" help:"Show package-level production and test metrics."`
	Files    filesCmd    `cmd:"" help:"Show file-level classification and line counts."`
	Tests    testsCmd    `cmd:"" help:"Summarize tests, subtests, benchmarks, and examples."`
	Version  versionCmd  `cmd:"" help:"Print the gocensus version."`
}

type analysisFlags struct {
	NoGitignore      bool     `name:"no-gitignore" help:"Do not read .gitignore exclude rules."`
	ExtraExcludes    []string `name:"exclude" short:"x" placeholder:"PATTERN" help:"Exclude paths matching pattern; can be repeated."`
	IncludeGenerated bool     `name:"include-generated" help:"Include generated files in production totals."`
	IncludeMocks     bool     `name:"include-mocks" help:"Include mock files in production totals."`
}

type rootArg struct {
	Root string `arg:"" optional:"" default:"." type:"path" help:"Repository root to analyze. Defaults to the current directory."`
}

type scanCmd struct {
	rootArg
	Format string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type reportCmd struct {
	rootArg
	Format string `short:"f" enum:"table,json,markdown" default:"markdown" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type packagesCmd struct {
	rootArg
	Sort string `short:"s" enum:"path,test-ratio,prod-lines,test-lines" default:"path" help:"Sort packages by path, test-ratio, prod-lines, or test-lines."`
}

type filesCmd struct {
	rootArg
	Top int `short:"n" default:"20" help:"Maximum number of files to print; use 0 for all files."`
}

type testsCmd struct {
	rootArg
}

type versionCmd struct{}

func (cmd scanCmd) Help() string {
	return "Analyze a Go repository and print the main census: files, raw and effective lines, production/test mix, ratios, known test cases, and notes for each metric."
}

func (cmd reportCmd) Help() string {
	return "Generate the same repository analysis as a report. Markdown is the default format so the output can be saved in CI artifacts, release notes, or project documentation."
}

func (cmd packagesCmd) Help() string {
	return "List each package with effective production and test lines plus its test-to-production ratio. Sort by test-ratio, prod-lines, or test-lines to find thinly tested or large packages."
}

func (cmd filesCmd) Help() string {
	return "List file-level classification and line counts. Use this to find large files or files that landed in an unexpected test, generated, or mock bucket."
}

func (cmd testsCmd) Help() string {
	return "Summarize test inventory: top-level tests, statically countable subtests, dynamic subtest sites, benchmarks, subbenchmarks, and examples."
}

func (cmd versionCmd) Help() string {
	return "Print the gocensus version and build identifier."
}

func (cmd *scanCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output)
}

func (cmd *reportCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output)
}

func (cmd *packagesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderPackages(rt.stdout, result, cmd.Sort)
}

func (cmd *filesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderFiles(rt.stdout, result, cmd.Top)
}

func (cmd *testsCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderTests(rt.stdout, result)
}

func (cmd *versionCmd) Run(rt *runtime) error {
	_, err := fmt.Fprintf(rt.stdout, "gocensus %s\n", rt.version)
	return err
}

func analyze(rt *runtime, root string, flags analysisFlags) (gocensus.Result, error) {
	return gocensus.Analyze(rt.ctx, gocensus.Options{
		Root:             root,
		NoGitignore:      flags.NoGitignore,
		ExtraExcludes:    flags.ExtraExcludes,
		IncludeGenerated: flags.IncludeGenerated,
		IncludeMocks:     flags.IncludeMocks,
	})
}
