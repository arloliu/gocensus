package cli

import (
	"bytes"
	"fmt"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/color"
	"github.com/arloliu/gocensus/internal/contrib"
	censusdiff "github.com/arloliu/gocensus/internal/diff"
	"github.com/arloliu/gocensus/internal/gitrepo"
	"github.com/arloliu/gocensus/internal/hotspot"
)

type commandLine struct {
	analysisFlags `embed:""`

	Color   string `name:"color" enum:"auto,always,never" default:"auto" help:"Control terminal colors: auto, always, or never."`
	NoColor bool   `name:"no-color" help:"Disable terminal colors."`

	Scan     scanCmd     `cmd:"" default:"withargs" help:"Print the main repository census."`
	Report   reportCmd   `cmd:"" help:"Write a repository report."`
	Packages packagesCmd `cmd:"" help:"Show package-level production and test metrics."`
	Files    filesCmd    `cmd:"" help:"Show file-level classification and line counts."`
	Tests    testsCmd    `cmd:"" help:"Summarize tests, subtests, benchmarks, and examples."`
	Who      whoCmd      `cmd:"" help:"Rank contributors from Git history."`
	Diff     diffCmd     `cmd:"" help:"Compare Go census metrics between two Git refs."`
	Hotspots hotspotsCmd `cmd:"" help:"Rank human-authored Go file hotspots by size and Git churn."`
	Version  versionCmd  `cmd:"" help:"Print the gocensus version."`
}

type analysisFlags struct {
	NoGitignore      bool     `name:"no-gitignore" help:"Do not read .gitignore exclude rules."`
	ExtraExcludes    []string `name:"exclude" short:"x" placeholder:"PATTERN" help:"Exclude paths matching pattern; can be repeated."`
	IncludeGenerated bool     `name:"include-generated" help:"Include generated files in scan production totals; with who --go-only, include generated Go paths."`
	IncludeMocks     bool     `name:"include-mocks" help:"Include mock files in scan production totals; with who --go-only, include mock Go paths."`
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

type whoCmd struct {
	rootArg
	GoOnly           bool   `name:"go-only" help:"Rank human-authored Go paths only: *.go with generated and mock paths excluded unless --include-generated or --include-mocks is set."`
	ExcludeGenerated bool   `name:"exclude-generated" help:"Exclude generated paths from contributor rankings across all tracked file types."`
	ExcludeMocks     bool   `name:"exclude-mocks" help:"Exclude mock paths from contributor rankings across all tracked file types."`
	By               string `name:"by" enum:"commits,features,fixes,refactors,added,removed,net,shrink,churn,files,active-days" default:"commits" help:"Rank by commits, features, fixes, refactors, added, removed, net, shrink, churn, files, or active-days."`
	Top              int    `short:"n" default:"10" help:"Maximum number of contributors to print; use 0 for all contributors."`
	Since            string `name:"since" placeholder:"DATE" help:"Only include commits after this date or Git revision expression."`
	Until            string `name:"until" placeholder:"DATE" help:"Only include commits before this date or Git revision expression."`
	Format           string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output           string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type diffCmd struct {
	rootArg
	Base   string `name:"base" default:"HEAD~1" placeholder:"REF" help:"Base Git ref to analyze."`
	Head   string `name:"head" default:"HEAD" placeholder:"REF" help:"Head Git ref to analyze."`
	Format string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}

type hotspotsCmd struct {
	rootArg
	By     string `name:"by" enum:"score,lines,churn,commits,test-ratio" default:"score" help:"Sort by score, lines, churn, commits, or test-ratio."`
	Top    int    `short:"n" default:"20" help:"Maximum number of files to print; use 0 for all files."`
	Since  string `name:"since" placeholder:"DATE" help:"Only include Git churn after this date or Git revision expression."`
	Until  string `name:"until" placeholder:"DATE" help:"Only include Git churn before this date or Git revision expression."`
	Format string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
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

func (cmd whoCmd) Help() string {
	return "Rank Git authors by commit count, message-classified feature/fix/refactor work, added and removed lines, net change, churn, files touched, and active days. By default this uses all Git-tracked files. Use --exclude-generated and --exclude-mocks to remove generated or mock paths across all tracked file types. Use --go-only for the recommended human-authored Go view; with --go-only, --include-generated and --include-mocks add generated or mock Go paths back into the result."
}

func (cmd diffCmd) Help() string {
	return "Compare scan metrics between two Git refs without mutating the working tree. Scope follows scan: generated and mock files are excluded from production totals unless --include-generated or --include-mocks is set."
}

func (cmd hotspotsCmd) Help() string {
	return "Rank human-authored production Go files by effective lines plus Git churn. Test files, generated files, and mock files are excluded by default; --include-generated and --include-mocks add those production-scope files back into the report."
}

func (cmd versionCmd) Help() string {
	return "Print the gocensus version and build identifier."
}

func (cmd *scanCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output, cli.style(rt))
}

func (cmd *reportCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderResult(rt.stdout, result, cmd.Format, cmd.Output, cli.style(rt))
}

func (cmd *packagesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderPackages(rt.stdout, result, cmd.Sort, cli.style(rt))
}

func (cmd *filesCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderFiles(rt.stdout, result, cmd.Top, cli.style(rt))
}

func (cmd *testsCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	return renderTests(rt.stdout, result, cli.style(rt))
}

func (cmd *whoCmd) Run(cli *commandLine, rt *runtime) error {
	report, err := contrib.Analyze(rt.ctx, contrib.Options{
		Root:             cmd.Root,
		Since:            cmd.Since,
		Until:            cmd.Until,
		GoOnly:           cmd.GoOnly,
		IncludeGenerated: cli.IncludeGenerated,
		IncludeMocks:     cli.IncludeMocks,
		ExcludeGenerated: cmd.ExcludeGenerated,
		ExcludeMocks:     cmd.ExcludeMocks,
	})
	if err != nil {
		return err
	}
	ranked, err := contrib.Rank(report, contrib.RankOptions{
		By:  cmd.By,
		Top: cmd.Top,
	})
	if err != nil {
		return err
	}
	report.Contributors = ranked
	return renderWho(rt.stdout, report, cmd.By, cmd.Format, cmd.Output, cli.style(rt))
}

func (cmd *diffCmd) Run(cli *commandLine, rt *runtime) error {
	baseRoot, cleanupBase, err := gitrepo.Archive(rt.ctx, cmd.Root, cmd.Base)
	if err != nil {
		return err
	}
	defer cleanupBase()
	headRoot, cleanupHead, err := gitrepo.Archive(rt.ctx, cmd.Root, cmd.Head)
	if err != nil {
		return err
	}
	defer cleanupHead()

	baseResult, err := analyze(rt, baseRoot, cli.analysisFlags)
	if err != nil {
		return err
	}
	headResult, err := analyze(rt, headRoot, cli.analysisFlags)
	if err != nil {
		return err
	}
	repoRoot, err := gitrepo.Root(rt.ctx, cmd.Root)
	if err != nil {
		return err
	}
	report := censusdiff.Compare(censusdiff.Options{
		Root:  repoRoot,
		Base:  cmd.Base,
		Head:  cmd.Head,
		Scope: headResult.Scope,
	}, baseResult, headResult)
	return renderDiff(rt.stdout, report, cmd.Format, cmd.Output, cli.style(rt))
}

func (cmd *hotspotsCmd) Run(cli *commandLine, rt *runtime) error {
	result, err := analyze(rt, cmd.Root, cli.analysisFlags)
	if err != nil {
		return err
	}
	numstat, err := gitrepo.Numstat(rt.ctx, cmd.Root, cmd.Since, cmd.Until)
	if err != nil {
		return err
	}
	churn, err := hotspot.ParseNumstat(bytes.NewReader(numstat))
	if err != nil {
		return err
	}
	report, err := hotspot.Rank(result, churn, hotspot.Options{
		By:               cmd.By,
		Top:              cmd.Top,
		IncludeGenerated: cli.IncludeGenerated,
		IncludeMocks:     cli.IncludeMocks,
	})
	if err != nil {
		return err
	}
	return renderHotspots(rt.stdout, report, cmd.Format, cmd.Output, cli.style(rt))
}

func (cmd *versionCmd) Run(rt *runtime) error {
	_, err := fmt.Fprintf(rt.stdout, "gocensus %s\n", rt.version)
	return err
}

func (cli *commandLine) style(rt *runtime) color.Style {
	return color.Resolve(color.Request{
		Mode:    cli.Color,
		NoColor: cli.NoColor,
		Environ: rt.environ,
	})
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
