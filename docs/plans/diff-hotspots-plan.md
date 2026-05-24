# Diff and Hotspots Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `gocensus diff` and `gocensus hotspots` so users can see how a Go repository changed between Git refs and which human-authored Go files deserve review attention.

**Architecture:** Keep analysis logic out of CLI command handlers. Add focused internal packages for Git snapshot materialization, census diffing, and hotspot scoring; CLI commands parse flags, call those packages, and render table/json/markdown through small view helpers. Reuse the existing scan scope contract: generated and mock files are excluded by default and included only when `--include-generated` or `--include-mocks` is passed.

**Tech Stack:** Go 1.21+, standard library, existing `gocensus.Analyze`, existing `internal/contrib` Git log parsing style, existing `internal/cli` text table renderer, `git archive`, `git log --numstat`, `go test ./...`, `make check`.

---

## Release Scope

Ship both commands in one release because they share Git helpers and scope language:

- `gocensus diff [root] --base REF --head REF`
- `gocensus hotspots [root]`

Default scopes:

- `diff` analyzes Go files with the same production scope as `scan`: generated and mock files are excluded from production totals by default.
- `hotspots` ranks human-authored production Go files by default: test files, generated files, and mock files are excluded unless the include flags move generated/mock files into production scope.
- Both command outputs print `Scope:` near the top.
- Both command helps explain whether generated and mock files are excluded or included.

Non-goals for this release:

- No HTML dashboard.
- No multi-language analysis.
- No AST complexity scoring.
- No working-tree pseudo-ref. `--base` and `--head` must be Git refs accepted by `git archive`.
- No persistent cache.

## File Structure

- Create `internal/gitrepo/gitrepo.go`: small Git command wrapper for repo root discovery, `git archive` extraction to temp dirs, and `git log --numstat` execution.
- Create `internal/gitrepo/gitrepo_test.go`: tests for command argument construction and archive extraction against a temp Git repo.
- Create `internal/diff/diff.go`: compare two `gocensus.Result` values and produce repository/package/file deltas.
- Create `internal/diff/diff_test.go`: unit tests for positive, negative, and zero deltas.
- Create `internal/hotspot/hotspot.go`: combine current file metrics with Git churn metrics and produce ranked file hotspots.
- Create `internal/hotspot/hotspot_test.go`: unit tests for filtering, scoring, sorting, and scope description.
- Modify `internal/cli/commands.go`: add `Diff` and `Hotspots` subcommands plus flags.
- Create `internal/cli/diff_view.go`: render diff table/json/markdown and write `-o/--output`.
- Create `internal/cli/hotspots_view.go`: render hotspots table/json/markdown and write `-o/--output`.
- Modify `internal/cli/help_test.go`: assert command help lists `diff` and `hotspots`, and command help mentions scope.
- Modify `internal/cli/cli_test.go`: add end-to-end CLI tests with temp Git repos.
- Modify `internal/cli/views_test.go`: add output-shape tests for diff and hotspots views.
- Modify `README.md`: document command purpose, examples, flags, scope semantics, and metric meanings.

## User-Facing Behavior

### `gocensus diff`

Command:

```bash
gocensus diff . --base HEAD~1 --head HEAD
```

Default flags:

```text
--base HEAD~1
--head HEAD
-f, --format table
-o, --output PATH
--include-generated false
--include-mocks false
```

Table shape:

```text
Diff: /repo
Base: HEAD~1
Head: HEAD
Scope: production excludes generated and mock files

Summary
  Metric                      Base     Head    Delta
  Go Files                      24       26       +2
  Production Files              14       16       +2
  Test Files                    10       10        0
  Production Effective       1,320    1,456     +136
  Test Effective               848      900      +52
  Test / Production Scope     0.64:1   0.62:1  -0.02
  Test Share                   39.1%    38.2%   -0.9%

Packages
  Package                  Prod Eff   Test Eff   Ratio   Prod Delta   Test Delta
  internal/cli                  420        280  0.67:1          +80          +20
```

### `gocensus hotspots`

Command:

```bash
gocensus hotspots . --since 90.days --by score -n 20
```

Default flags:

```text
--by score
--since ""
--until ""
-n, --top 20
-f, --format table
-o, --output PATH
--include-generated false
--include-mocks false
```

Score contract:

```text
Hotspot Score = Effective Lines + Git Churn
Git Churn = Added Lines + Removed Lines from git log --numstat
```

Table shape:

```text
Hotspots: /repo
Scope: human-authored production Go files (*.go, tests/generated/mock paths excluded)
Sorted by: score

  File                         Score  Eff Lines  Churn  Added  Removed  Commits  Package Test Ratio
  internal/cli/commands.go     1,024        420    604    400      204       18             0.67:1
```

Notes:

```text
Notes:
  Hotspot Score: effective production lines + Git churn.
  Git Churn: added + removed lines from git log --numstat.
  Package Test Ratio: package test effective lines divided by package production effective lines.
```

## Task 1: Add Git Repository Helpers

**Files:**
- Create: `internal/gitrepo/gitrepo.go`
- Create: `internal/gitrepo/gitrepo_test.go`

- [ ] **Step 1: Write failing tests for Git root and archive extraction**

Create `internal/gitrepo/gitrepo_test.go`:

```go
package gitrepo_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus/internal/gitrepo"
)

func TestRootReturnsTopLevelGitDirectory(t *testing.T) {
	dir := initRepo(t)
	nested := filepath.Join(dir, "internal", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := gitrepo.Root(context.Background(), nested)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Fatalf("root = %q, want %q", root, dir)
	}
}

func TestArchiveExtractsRefWithoutMutatingWorktree(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "feat: initial")
	runGit(t, dir, "tag", "v1")
	writeFile(t, dir, "main.go", "package main\n\nfunc New() {}\n")

	out, cleanup, err := gitrepo.Archive(context.Background(), dir, "v1")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	content, err := os.ReadFile(filepath.Join(out, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "package main\n" {
		t.Fatalf("archived main.go = %q", string(content))
	}
	current, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "package main\n\nfunc New() {}\n" {
		t.Fatalf("worktree main.go = %q", string(current))
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
	return dir
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func commitGit(t *testing.T, root string, message string) {
	t.Helper()
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", message)
}
```

- [ ] **Step 2: Run the new tests and confirm failure**

Run:

```bash
go test ./internal/gitrepo
```

Expected: FAIL because `internal/gitrepo` does not exist.

- [ ] **Step 3: Implement Git helpers**

Create `internal/gitrepo/gitrepo.go`:

```go
package gitrepo

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Root(ctx context.Context, root string) (string, error) {
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", "-C", absRoot, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", gitError("git rev-parse", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func Archive(ctx context.Context, root string, ref string) (string, func(), error) {
	if ref == "" {
		return "", nil, errors.New("git ref is required")
	}
	repoRoot, err := Root(ctx, root)
	if err != nil {
		return "", nil, err
	}
	out, err := os.MkdirTemp("", "gocensus-archive-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() {
		_ = os.RemoveAll(out)
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "archive", "--format=tar", ref)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if err := cmd.Start(); err != nil {
		cleanup()
		return "", nil, gitError("git archive", err)
	}
	if err := extractTar(pipe, out); err != nil {
		_ = cmd.Wait()
		cleanup()
		return "", nil, err
	}
	if err := cmd.Wait(); err != nil {
		cleanup()
		return "", nil, gitError("git archive", err)
	}
	return out, cleanup, nil
}

func Numstat(ctx context.Context, root string, since string, until string) ([]byte, error) {
	repoRoot, err := Root(ctx, root)
	if err != nil {
		return nil, err
	}
	args := []string{"-C", repoRoot, "log", "--numstat", "--date=short", "--format=format:%x1e%H%x1f%ad"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	args = append(args, "--", ".")
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, gitError("git log", err)
	}
	return output, nil
}

func extractTar(r io.Reader, out string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(out, filepath.Clean(header.Name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tr); err != nil {
				_ = file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
	}
}

func gitError(action string, err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("%s: %s", action, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return fmt.Errorf("%s: %w", action, err)
}
```

- [ ] **Step 4: Run Git helper tests**

Run:

```bash
go test ./internal/gitrepo
```

Expected: PASS.

## Task 2: Add Diff Analysis Package

**Files:**
- Create: `internal/diff/diff.go`
- Create: `internal/diff/diff_test.go`

- [ ] **Step 1: Write failing tests for result deltas**

Create `internal/diff/diff_test.go`:

```go
package diff_test

import (
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/diff"
)

func TestCompareComputesRepositoryDeltas(t *testing.T) {
	base := gocensus.Result{
		Files: gocensus.FileCounts{Total: 2, Production: 1, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Effective: 100},
			Tests:      gocensus.Metric{Effective: 50},
		},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
	}
	head := gocensus.Result{
		Files: gocensus.FileCounts{Total: 3, Production: 2, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Effective: 160},
			Tests:      gocensus.Metric{Effective: 80},
		},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
	}

	report := diff.Compare(diff.Options{
		Root:  "/repo",
		Base:  "v1",
		Head:  "v2",
		Scope: "production excludes generated and mock files",
	}, base, head)

	if report.Root != "/repo" || report.Base != "v1" || report.Head != "v2" {
		t.Fatalf("identity = %#v", report)
	}
	if report.Summary.TotalFiles.Delta != 1 {
		t.Fatalf("total delta = %d, want 1", report.Summary.TotalFiles.Delta)
	}
	if report.Summary.ProductionEffective.Delta != 60 {
		t.Fatalf("production effective delta = %d, want 60", report.Summary.ProductionEffective.Delta)
	}
	if report.Summary.TestEffective.Delta != 30 {
		t.Fatalf("test effective delta = %d, want 30", report.Summary.TestEffective.Delta)
	}
}
```

- [ ] **Step 2: Run diff package tests and confirm failure**

Run:

```bash
go test ./internal/diff
```

Expected: FAIL because `internal/diff` does not exist.

- [ ] **Step 3: Implement delta types and comparison**

Create `internal/diff/diff.go`:

```go
package diff

import (
	"slices"

	"github.com/arloliu/gocensus"
)

type Options struct {
	Root  string
	Base  string
	Head  string
	Scope string
}

type Report struct {
	Root     string         `json:"root"`
	Base     string         `json:"base"`
	Head     string         `json:"head"`
	Scope    string         `json:"scope"`
	Summary  Summary        `json:"summary"`
	Packages []PackageDelta `json:"packages"`
}

type Summary struct {
	TotalFiles          IntDelta   `json:"total_files"`
	ProductionFiles     IntDelta   `json:"production_files"`
	TestFiles           IntDelta   `json:"test_files"`
	ProductionEffective IntDelta   `json:"production_effective"`
	TestEffective       IntDelta   `json:"test_effective"`
	TestToProduction    FloatDelta `json:"test_to_production"`
	TestShare           FloatDelta `json:"test_share"`
}

type PackageDelta struct {
	Package             string     `json:"package"`
	ProductionEffective IntDelta   `json:"production_effective"`
	TestEffective       IntDelta   `json:"test_effective"`
	TestToProduction    FloatDelta `json:"test_to_production"`
}

type IntDelta struct {
	Base  int `json:"base"`
	Head  int `json:"head"`
	Delta int `json:"delta"`
}

type FloatDelta struct {
	Base  float64 `json:"base"`
	Head  float64 `json:"head"`
	Delta float64 `json:"delta"`
}

func Compare(opts Options, base gocensus.Result, head gocensus.Result) Report {
	return Report{
		Root:  opts.Root,
		Base:  opts.Base,
		Head:  opts.Head,
		Scope: opts.Scope,
		Summary: Summary{
			TotalFiles:          intDelta(base.Files.Total, head.Files.Total),
			ProductionFiles:     intDelta(base.Files.Production, head.Files.Production),
			TestFiles:           intDelta(base.Files.Tests, head.Files.Tests),
			ProductionEffective: intDelta(base.Lines.Production.Effective, head.Lines.Production.Effective),
			TestEffective:       intDelta(base.Lines.Tests.Effective, head.Lines.Tests.Effective),
			TestToProduction:    floatDelta(base.Ratios.TestToProductionEffective, head.Ratios.TestToProductionEffective),
			TestShare:           floatDelta(base.Ratios.TestShareEffective, head.Ratios.TestShareEffective),
		},
		Packages: comparePackages(base.Packages, head.Packages),
	}
}

func intDelta(base int, head int) IntDelta {
	return IntDelta{Base: base, Head: head, Delta: head - base}
}

func floatDelta(base float64, head float64) FloatDelta {
	return FloatDelta{Base: base, Head: head, Delta: head - base}
}

func comparePackages(base []gocensus.PackageMetric, head []gocensus.PackageMetric) []PackageDelta {
	byPackage := map[string]PackageDelta{}
	for _, pkg := range base {
		delta := byPackage[pkg.ImportPath]
		delta.Package = pkg.ImportPath
		delta.ProductionEffective.Base = pkg.Lines.Production.Effective
		delta.TestEffective.Base = pkg.Lines.Tests.Effective
		delta.TestToProduction.Base = pkg.Ratios.TestToProductionEffective
		byPackage[pkg.ImportPath] = delta
	}
	for _, pkg := range head {
		delta := byPackage[pkg.ImportPath]
		delta.Package = pkg.ImportPath
		delta.ProductionEffective.Head = pkg.Lines.Production.Effective
		delta.TestEffective.Head = pkg.Lines.Tests.Effective
		delta.TestToProduction.Head = pkg.Ratios.TestToProductionEffective
		delta.ProductionEffective.Delta = delta.ProductionEffective.Head - delta.ProductionEffective.Base
		delta.TestEffective.Delta = delta.TestEffective.Head - delta.TestEffective.Base
		delta.TestToProduction.Delta = delta.TestToProduction.Head - delta.TestToProduction.Base
		byPackage[pkg.ImportPath] = delta
	}
	packages := make([]PackageDelta, 0, len(byPackage))
	for _, delta := range byPackage {
		if delta.ProductionEffective.Delta == 0 && delta.TestEffective.Delta == 0 && delta.TestToProduction.Delta == 0 {
			continue
		}
		packages = append(packages, delta)
	}
	slices.SortFunc(packages, func(a, b PackageDelta) int {
		aAbs := abs(a.ProductionEffective.Delta)
		bAbs := abs(b.ProductionEffective.Delta)
		if aAbs > bAbs {
			return -1
		}
		if aAbs < bAbs {
			return 1
		}
		if a.Package < b.Package {
			return -1
		}
		if a.Package > b.Package {
			return 1
		}
		return 0
	})
	return packages
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
```

- [ ] **Step 4: Run diff package tests**

Run:

```bash
go test ./internal/diff
```

Expected: PASS.

## Task 3: Add Diff CLI and Renderers

**Files:**
- Modify: `internal/cli/commands.go`
- Create: `internal/cli/diff_view.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/views_test.go`
- Modify: `internal/cli/help_test.go`

- [ ] **Step 1: Write CLI tests for `gocensus diff`**

Add this test to `internal/cli/cli_test.go`:

```go
func TestRunDiffComparesGitRefs(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	runGit(t, dir, "tag", "base")
	writeFile(t, dir, "main.go", "package main\n\nfunc NewFeature() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author Two", "two@example.com", "2026-01-04T12:00:00Z", "feat: expand")
	runGit(t, dir, "tag", "head")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"diff", dir, "--base", "base", "--head", "head"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Diff:", "Base: base", "Head: head", "Scope:", "Production Effective", "+"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}
```

- [ ] **Step 2: Run the CLI test and confirm failure**

Run:

```bash
go test ./internal/cli -run TestRunDiffComparesGitRefs
```

Expected: FAIL because `diff` is not a registered command.

- [ ] **Step 3: Add command shape**

In `internal/cli/commands.go`, add the command field:

```go
Diff diffCmd `cmd:"" help:"Compare Go census metrics between two Git refs."`
```

Add the command type:

```go
type diffCmd struct {
	rootArg
	Base   string `name:"base" default:"HEAD~1" placeholder:"REF" help:"Base Git ref to analyze."`
	Head   string `name:"head" default:"HEAD" placeholder:"REF" help:"Head Git ref to analyze."`
	Format string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}
```

Add help text:

```go
func (cmd diffCmd) Help() string {
	return "Compare scan metrics between two Git refs without mutating the working tree. Scope follows scan: generated and mock files are excluded from production totals unless --include-generated or --include-mocks is set."
}
```

- [ ] **Step 4: Implement command execution**

In `internal/cli/commands.go`, add imports for `internal/diff` and `internal/gitrepo`, then add:

```go
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
	report := diff.Compare(diff.Options{
		Root:  repoRoot,
		Base:  cmd.Base,
		Head:  cmd.Head,
		Scope: headResult.Scope,
	}, baseResult, headResult)
	return renderDiff(rt.stdout, report, cmd.Format, cmd.Output)
}
```

- [ ] **Step 5: Create diff renderers**

Create `internal/cli/diff_view.go` with:

```go
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/arloliu/gocensus/internal/diff"
)

func renderDiff(w io.Writer, report diff.Report, format string, output string) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	switch format {
	case "table":
		if err := renderDiffTable(writer, report); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderDiffMarkdown(writer, report); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderDiffTable(w io.Writer, report diff.Report) error {
	if _, err := fmt.Fprintf(w, "Diff: %s\n", report.Root); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Base: %s\n", report.Base); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Head: %s\n", report.Head); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "\nSummary"); err != nil {
		return err
	}
	table := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: "Metric", Align: tableAlignLeft},
			{Header: "Base", Align: tableAlignRight},
			{Header: "Head", Align: tableAlignRight},
			{Header: "Delta", Align: tableAlignRight},
		},
		Rows: diffSummaryRows(report.Summary),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Packages) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\nPackages"); err != nil {
		return err
	}
	packageTable := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: "Package", Align: tableAlignLeft},
			{Header: "Prod Eff", Align: tableAlignRight},
			{Header: "Test Eff", Align: tableAlignRight},
			{Header: "Ratio", Align: tableAlignRight},
			{Header: "Prod Delta", Align: tableAlignRight},
			{Header: "Test Delta", Align: tableAlignRight},
		},
		Rows: diffPackageRows(report.Packages),
	}
	return renderTextTable(w, packageTable)
}

func renderDiffMarkdown(w io.Writer, report diff.Report) error {
	if _, err := fmt.Fprintf(w, "# Diff: %s\n\n", report.Root); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Base: `%s`\n\nHead: `%s`\n\n", report.Base, report.Head); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "| Metric | Base | Head | Delta |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, row := range diffSummaryRows(report.Summary) {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3]); err != nil {
			return err
		}
	}
	return nil
}

func diffSummaryRows(summary diff.Summary) [][]string {
	return [][]string{
		{"Go Files", formatWhoInt(summary.TotalFiles.Base), formatWhoInt(summary.TotalFiles.Head), signedInt(summary.TotalFiles.Delta)},
		{"Production Files", formatWhoInt(summary.ProductionFiles.Base), formatWhoInt(summary.ProductionFiles.Head), signedInt(summary.ProductionFiles.Delta)},
		{"Test Files", formatWhoInt(summary.TestFiles.Base), formatWhoInt(summary.TestFiles.Head), signedInt(summary.TestFiles.Delta)},
		{"Production Effective", formatWhoInt(summary.ProductionEffective.Base), formatWhoInt(summary.ProductionEffective.Head), signedInt(summary.ProductionEffective.Delta)},
		{"Test Effective", formatWhoInt(summary.TestEffective.Base), formatWhoInt(summary.TestEffective.Head), signedInt(summary.TestEffective.Delta)},
		{"Test / Production Scope", formatDiffRatio(summary.TestToProduction.Base), formatDiffRatio(summary.TestToProduction.Head), signedFloat(summary.TestToProduction.Delta)},
		{"Test Share", formatDiffPercent(summary.TestShare.Base), formatDiffPercent(summary.TestShare.Head), signedPercent(summary.TestShare.Delta)},
	}
}

func diffPackageRows(packages []diff.PackageDelta) [][]string {
	rows := make([][]string, 0, len(packages))
	for _, pkg := range packages {
		rows = append(rows, []string{
			pkg.Package,
			formatWhoInt(pkg.ProductionEffective.Head),
			formatWhoInt(pkg.TestEffective.Head),
			formatDiffRatio(pkg.TestToProduction.Head),
			signedInt(pkg.ProductionEffective.Delta),
			signedInt(pkg.TestEffective.Delta),
		})
	}
	return rows
}

func signedInt(value int) string {
	if value > 0 {
		return "+" + formatWhoInt(value)
	}
	return formatWhoInt(value)
}

func formatDiffRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}

func formatDiffPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func signedFloat(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.2f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func signedPercent(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.1f%%", value*100)
	}
	return fmt.Sprintf("%.1f%%", value*100)
}
```

- [ ] **Step 6: Run CLI diff tests**

Run:

```bash
go test ./internal/cli -run 'TestRunDiff|TestRenderDiff'
```

Expected: PASS.

## Task 4: Add Hotspot Analysis Package

**Files:**
- Create: `internal/hotspot/hotspot.go`
- Create: `internal/hotspot/hotspot_test.go`

- [ ] **Step 1: Write failing tests for scoring and filtering**

Create `internal/hotspot/hotspot_test.go`:

```go
package hotspot_test

import (
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/hotspot"
)

func TestRankFiltersToProductionFilesAndScoresWithChurn(t *testing.T) {
	result := gocensus.Result{
		Root:  "/repo",
		Scope: "production excludes generated and mock files",
		FileMetrics: []gocensus.FileMetric{
			{Path: "main.go", Package: "example.com/app", Kind: "production", CodeLines: 100},
			{Path: "main_test.go", Package: "example.com/app", Kind: "test", CodeLines: 200},
			{Path: "service.pb.go", Package: "example.com/app", Kind: "generated", CodeLines: 300},
		},
		Packages: []gocensus.PackageMetric{
			{
				ImportPath: "example.com/app",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 100},
					Tests:      gocensus.Metric{Effective: 50},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.5},
			},
		},
	}
	churn := map[string]hotspot.Churn{
		"main.go": {Added: 70, Removed: 30, Commits: 3},
	}

	report, err := hotspot.Rank(result, churn, hotspot.Options{By: "score", Top: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 {
		t.Fatalf("file count = %d, want 1", len(report.Files))
	}
	got := report.Files[0]
	if got.Path != "main.go" {
		t.Fatalf("path = %q, want main.go", got.Path)
	}
	if got.Score != 200 {
		t.Fatalf("score = %d, want 200", got.Score)
	}
	if got.Churn != 100 {
		t.Fatalf("churn = %d, want 100", got.Churn)
	}
	if got.PackageTestRatio != 0.5 {
		t.Fatalf("package test ratio = %v, want 0.5", got.PackageTestRatio)
	}
}
```

- [ ] **Step 2: Run hotspot package tests and confirm failure**

Run:

```bash
go test ./internal/hotspot
```

Expected: FAIL because `internal/hotspot` does not exist.

- [ ] **Step 3: Implement hotspot types and ranking**

Create `internal/hotspot/hotspot.go`:

```go
package hotspot

import (
	"fmt"
	"slices"

	"github.com/arloliu/gocensus"
)

const (
	SortScore   = "score"
	SortLines   = "lines"
	SortChurn   = "churn"
	SortCommits = "commits"
	SortRatio   = "test-ratio"
)

type Options struct {
	By  string
	Top int
}

type Churn struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Commits int `json:"commits"`
}

type Report struct {
	Root   string   `json:"root"`
	Scope  string   `json:"scope"`
	SortBy string   `json:"sort_by"`
	Notes  []string `json:"notes,omitempty"`
	Files  []File   `json:"files"`
}

type File struct {
	Path             string  `json:"path"`
	Package          string  `json:"package"`
	Score            int     `json:"score"`
	EffectiveLines   int     `json:"effective_lines"`
	Churn            int     `json:"churn"`
	Added            int     `json:"added"`
	Removed          int     `json:"removed"`
	Commits          int     `json:"commits"`
	PackageTestRatio float64 `json:"package_test_ratio"`
}

func Rank(result gocensus.Result, churnByPath map[string]Churn, opts Options) (Report, error) {
	if opts.By == "" {
		opts.By = SortScore
	}
	if !validSort(opts.By) {
		return Report{}, fmt.Errorf("unknown hotspot sort %q", opts.By)
	}
	ratios := packageRatios(result.Packages)
	files := make([]File, 0, len(result.FileMetrics))
	for _, metric := range result.FileMetrics {
		if metric.Kind != "production" {
			continue
		}
		churn := churnByPath[metric.Path]
		totalChurn := churn.Added + churn.Removed
		files = append(files, File{
			Path:             metric.Path,
			Package:          metric.Package,
			Score:            metric.CodeLines + totalChurn,
			EffectiveLines:   metric.CodeLines,
			Churn:            totalChurn,
			Added:            churn.Added,
			Removed:          churn.Removed,
			Commits:          churn.Commits,
			PackageTestRatio: ratios[metric.Package],
		})
	}
	sortFiles(files, opts.By)
	if opts.Top > 0 && len(files) > opts.Top {
		files = files[:opts.Top]
	}
	return Report{
		Root:   result.Root,
		Scope:  result.Scope,
		SortBy: opts.By,
		Notes: []string{
			"Hotspot Score = effective production lines + Git churn.",
			"Git Churn = added + removed lines from git log --numstat.",
			"Package Test Ratio = package test effective lines divided by package production effective lines.",
		},
		Files: files,
	}, nil
}
func validSort(sortBy string) bool {
	switch sortBy {
	case SortScore, SortLines, SortChurn, SortCommits, SortRatio:
		return true
	default:
		return false
	}
}

func packageRatios(packages []gocensus.PackageMetric) map[string]float64 {
	ratios := make(map[string]float64, len(packages))
	for _, pkg := range packages {
		ratios[pkg.ImportPath] = pkg.Ratios.TestToProductionEffective
	}
	return ratios
}

func sortFiles(files []File, sortBy string) {
	slices.SortFunc(files, func(a, b File) int {
		switch sortBy {
		case SortScore:
			if a.Score != b.Score {
				return descending(a.Score, b.Score)
			}
		case SortLines:
			if a.EffectiveLines != b.EffectiveLines {
				return descending(a.EffectiveLines, b.EffectiveLines)
			}
		case SortChurn:
			if a.Churn != b.Churn {
				return descending(a.Churn, b.Churn)
			}
		case SortCommits:
			if a.Commits != b.Commits {
				return descending(a.Commits, b.Commits)
			}
		case SortRatio:
			if a.PackageTestRatio > b.PackageTestRatio {
				return -1
			}
			if a.PackageTestRatio < b.PackageTestRatio {
				return 1
			}
		}
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})
}

func descending(a int, b int) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}
```

- [ ] **Step 4: Run hotspot package tests**

Run:

```bash
go test ./internal/hotspot
```

Expected: PASS.

## Task 5: Parse Churn for Hotspots

**Files:**
- Modify: `internal/hotspot/hotspot.go`
- Modify: `internal/hotspot/hotspot_test.go`

- [ ] **Step 1: Write failing parser test**

Add this test to `internal/hotspot/hotspot_test.go`:

```go
func TestParseNumstatAggregatesChurnByPath(t *testing.T) {
	input := strings.NewReader("\x1eabc\x1f2026-01-03\n10\t2\tmain.go\n-\t-\tasset.bin\n\x1edef\x1f2026-01-04\n3\t4\tmain.go\n5\t0\tpkg/service.go\n")

	churn, err := hotspot.ParseNumstat(input)
	if err != nil {
		t.Fatal(err)
	}
	if churn["main.go"].Added != 13 || churn["main.go"].Removed != 6 || churn["main.go"].Commits != 2 {
		t.Fatalf("main.go churn = %#v", churn["main.go"])
	}
	if churn["pkg/service.go"].Added != 5 || churn["pkg/service.go"].Removed != 0 || churn["pkg/service.go"].Commits != 1 {
		t.Fatalf("pkg/service.go churn = %#v", churn["pkg/service.go"])
	}
	if _, ok := churn["asset.bin"]; ok {
		t.Fatalf("binary file should be skipped: %#v", churn["asset.bin"])
	}
}
```

Add `strings` to the test imports.

- [ ] **Step 2: Run parser test and confirm failure**

Run:

```bash
go test ./internal/hotspot -run TestParseNumstatAggregatesChurnByPath
```

Expected: FAIL because `ParseNumstat` does not exist.

- [ ] **Step 3: Implement parser**

Add to `internal/hotspot/hotspot.go`:

```go
func ParseNumstat(r io.Reader) (map[string]Churn, error) {
	scanner := bufio.NewScanner(r)
	out := map[string]Churn{}
	seenInCommit := map[string]struct{}{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "\x1e") {
			seenInCommit = map[string]struct{}{}
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 || fields[0] == "-" || fields[1] == "-" {
			continue
		}
		added, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, err
		}
		removed, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}
		path := filepath.ToSlash(fields[2])
		churn := out[path]
		churn.Added += added
		churn.Removed += removed
		if _, ok := seenInCommit[path]; !ok {
			churn.Commits++
			seenInCommit[path] = struct{}{}
		}
		out[path] = churn
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
```

Add imports for `bufio`, `io`, `path/filepath`, `strconv`, and `strings`.

- [ ] **Step 4: Run hotspot tests**

Run:

```bash
go test ./internal/hotspot
```

Expected: PASS.

## Task 6: Add Hotspots CLI and Renderers

**Files:**
- Modify: `internal/cli/commands.go`
- Create: `internal/cli/hotspots_view.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/views_test.go`
- Modify: `internal/cli/help_test.go`

- [ ] **Step 1: Write CLI test for `gocensus hotspots`**

Add this test to `internal/cli/cli_test.go`:

```go
func TestRunHotspotsRanksProductionFiles(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "small.go", "package main\n")
	writeFile(t, dir, "large.go", "package main\n\nfunc One() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	writeFile(t, dir, "large.go", "package main\n\nfunc One() {}\n\nfunc Two() {}\n\nfunc Three() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author Two", "two@example.com", "2026-01-04T12:00:00Z", "feat: expand")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"hotspots", dir, "-n", "1"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Hotspots:", "Scope:", "Sorted by: score", "Score", "Eff Lines", "Churn", "large.go"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	if strings.Contains(output, "small.go") {
		t.Fatalf("stdout = %q, did not want small.go when top is 1", output)
	}
}
```

- [ ] **Step 2: Run the CLI test and confirm failure**

Run:

```bash
go test ./internal/cli -run TestRunHotspotsRanksProductionFiles
```

Expected: FAIL because `hotspots` is not registered.

- [ ] **Step 3: Add command shape**

In `internal/cli/commands.go`, add the command field:

```go
Hotspots hotspotsCmd `cmd:"" help:"Rank human-authored Go file hotspots by size and Git churn."`
```

Add the command type:

```go
type hotspotsCmd struct {
	rootArg
	By     string `name:"by" enum:"score,lines,churn,commits,test-ratio" default:"score" help:"Sort by score, lines, churn, commits, or test-ratio."`
	Top    int    `short:"n" default:"20" help:"Maximum number of files to print; use 0 for all files."`
	Since  string `name:"since" placeholder:"DATE" help:"Only include Git churn after this date or Git revision expression."`
	Until  string `name:"until" placeholder:"DATE" help:"Only include Git churn before this date or Git revision expression."`
	Format string `short:"f" enum:"table,json,markdown" default:"table" help:"Output format: table, json, or markdown."`
	Output string `short:"o" placeholder:"PATH" help:"Write output to file instead of stdout."`
}
```

Add help text:

```go
func (cmd hotspotsCmd) Help() string {
	return "Rank human-authored production Go files by effective lines plus Git churn. Test files, generated files, and mock files are excluded by default; --include-generated and --include-mocks add those production-scope files back into the report."
}
```

- [ ] **Step 4: Implement command execution**

In `internal/cli/commands.go`, add imports for `bytes`, `internal/gitrepo`, and `internal/hotspot` if they are not already present. Then add:

```go
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
	report, err := hotspot.Rank(result, churn, hotspot.Options{By: cmd.By, Top: cmd.Top})
	if err != nil {
		return err
	}
	return renderHotspots(rt.stdout, report, cmd.Format, cmd.Output)
}
```

- [ ] **Step 5: Create hotspots renderers**

Create `internal/cli/hotspots_view.go`:

```go
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/arloliu/gocensus/internal/hotspot"
)

func renderHotspots(w io.Writer, report hotspot.Report, format string, output string) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	switch format {
	case "table":
		if err := renderHotspotsTable(writer, report); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderHotspotsMarkdown(writer, report); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderHotspotsTable(w io.Writer, report hotspot.Report) error {
	if _, err := fmt.Fprintf(w, "Hotspots: %s\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: %s\n\n", report.SortBy); err != nil {
		return err
	}
	table := textTable{
		Indent:  "  ",
		Gap:     "  ",
		Columns: hotspotsTableColumns(),
		Rows:    hotspotsRows(report.Files),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Notes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\nNotes:"); err != nil {
		return err
	}
	for _, note := range report.Notes {
		if _, err := fmt.Fprintf(w, "  %s\n", note); err != nil {
			return err
		}
	}
	return nil
}

func hotspotsTableColumns() []tableColumn {
	return []tableColumn{
		{Header: "File", Align: tableAlignLeft},
		{Header: "Score", Align: tableAlignRight},
		{Header: "Eff Lines", Align: tableAlignRight},
		{Header: "Churn", Align: tableAlignRight},
		{Header: "Added", Align: tableAlignRight},
		{Header: "Removed", Align: tableAlignRight},
		{Header: "Commits", Align: tableAlignRight},
		{Header: "Package Test Ratio", Align: tableAlignRight},
	}
}

func hotspotsRows(files []hotspot.File) [][]string {
	rows := make([][]string, 0, len(files))
	for _, file := range files {
		rows = append(rows, []string{
			file.Path,
			formatWhoInt(file.Score),
			formatWhoInt(file.EffectiveLines),
			formatWhoInt(file.Churn),
			formatWhoInt(file.Added),
			formatWhoInt(file.Removed),
			formatWhoInt(file.Commits),
			formatHotspotRatio(file.PackageTestRatio),
		})
	}
	return rows
}

func renderHotspotsMarkdown(w io.Writer, report hotspot.Report) error {
	if _, err := fmt.Fprintf(w, "# Hotspots: %s\n\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: `%s`\n\n", report.SortBy); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| File | Score | Eff Lines | Churn | Added | Removed | Commits | Package Test Ratio |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, row := range hotspotsRows(report.Files) {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7]); err != nil {
			return err
		}
	}
	if len(report.Notes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\n## Notes"); err != nil {
		return err
	}
	for _, note := range report.Notes {
		if _, err := fmt.Fprintf(w, "- %s\n", note); err != nil {
			return err
		}
	}
	return nil
}

func formatHotspotRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
```

- [ ] **Step 6: Run CLI hotspots tests**

Run:

```bash
go test ./internal/cli -run 'TestRunHotspots|TestRenderHotspots'
```

Expected: PASS.

## Task 7: Update Help, README, and Output Contracts

**Files:**
- Modify: `README.md`
- Modify: `internal/cli/help_test.go`
- Modify: `internal/cli/views_test.go`

- [ ] **Step 1: Extend help tests**

In `internal/cli/help_test.go`, update the top-level help assertion list to include:

```go
"diff",
"hotspots",
```

Add command-help tests that assert:

```go
[]string{
	"generated",
	"mock",
	"--include-generated",
	"--include-mocks",
}
```

for both `diff --help` and `hotspots --help`.

- [ ] **Step 2: Add view tests for self-explaining headers**

In `internal/cli/views_test.go`, add assertions that rendered table output contains:

```go
"Scope:"
"Base:"
"Head:"
"Hotspot Score"
"Git Churn"
"Package Test Ratio"
```

Use synthetic `diff.Report` and `hotspot.Report` values with names like `example.com/app`, `internal/app/app.go`, and `Author One`; do not use real contributor names from local repositories.

- [ ] **Step 3: Update README commands table**

Add rows:

```markdown
| `gocensus diff [root]` | Compare scan metrics between two Git refs. |
| `gocensus hotspots [root]` | Rank human-authored Go file hotspots by size and Git churn. |
```

Add examples:

```bash
gocensus diff . --base v0.1.0 --head HEAD
gocensus diff . --base main --head feature/my-change -f markdown -o diff.md
gocensus hotspots . --since 90.days
gocensus hotspots . --by churn -n 20
```

Add metric meanings:

```markdown
| Hotspot Score | Effective production lines plus Git churn. |
| Git Churn | Added plus removed lines from `git log --numstat`. |
| Package Test Ratio | Package test effective lines divided by package production effective lines. |
| Diff Delta | Head value minus base value. Positive values are shown with `+` in table output. |
```

- [ ] **Step 4: Run focused doc/help tests**

Run:

```bash
go test ./internal/cli -run 'TestHelp|TestRender'
```

Expected: PASS.

## Task 8: Full Validation and Release Notes

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Run all tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run full project gate**

Run:

```bash
make check
```

Expected: PASS, including formatting, lint, tests, and CLI build.

- [ ] **Step 3: Manually verify commands on this repository**

Run:

```bash
go run ./cmd/gocensus diff . --base HEAD~1 --head HEAD
go run ./cmd/gocensus hotspots . -n 10
go run ./cmd/gocensus hotspots . --include-generated --include-mocks -n 10
```

Expected:

- `diff` prints `Diff:`, `Base:`, `Head:`, `Scope:`, `Summary`, and `Packages`.
- default `hotspots` prints `Scope:` with generated and mock paths excluded.
- `--include-generated --include-mocks` changes the scope line and allows those production-scope files into the report.

- [ ] **Step 4: Commit**

Run:

```bash
git add README.md internal/gitrepo internal/diff internal/hotspot internal/cli
git commit -m "feat: add diff and hotspots commands"
```

Expected: commit succeeds after `make check`. Commit message has no attribution trailers.

## Self-Review Checklist

- `diff` and `hotspots` both print scope near the top.
- Generated/mock inclusion flags affect totals and rankings, not only labels.
- Table/json/markdown formats exist for both commands.
- Git helpers do not mutate the working tree.
- Hotspot score formula is visible in output notes and README.
- Tests avoid real names from local repositories.
- All new code stays compatible with Go 1.21.
- `make check` passes before completion is claimed.
