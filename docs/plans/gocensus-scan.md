# Go Census Scan Implementation Plan

> **For implementers:** Execute this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the v1 `gocensus scan` analyzer as a reusable Go library plus CLI that reports production, test, generated, mock, package, file, and test declaration metrics.

**Architecture:** Keep analysis independent from presentation. The public `gocensus.Analyze(ctx, Options)` API delegates to internal packages for discovery, classification, counting, aggregation, and rendering; CLI commands only parse flags, call the API, and write renderer output.

**Tech Stack:** Go 1.21+, standard library `go/scanner`, `go/parser`, `go/ast`, `filepath.WalkDir`, `path.Match`, `slices`, `golangci-lint` v2, and Makefile.

---

## File Structure

- Modify `census.go`: expand public API types, options, ratios, package metrics, and file metrics.
- Modify `analyze_test.go`: cover end-to-end analysis behavior against fixture repositories.
- Create `internal/discover/discover.go`: walk repositories, apply hard excludes and `.gitignore`, return Go source files.
- Create `internal/discover/discover_test.go`: verify default excludes, root `.gitignore`, nested `.gitignore`, and `--no-gitignore` behavior.
- Create `internal/classify/classify.go`: classify files as production, test, generated, mock.
- Create `internal/classify/classify_test.go`: verify suffix, generated marker, protobuf, and mock classification.
- Create `internal/count/count.go`: count raw lines, effective code lines, and test declarations.
- Create `internal/count/count_test.go`: verify block comments, line comments, inline comments, blank lines, and `Test`/`Benchmark`/`Example` discovery.
- Create `internal/report/report.go`: aggregate file metrics into repo and package metrics.
- Create `internal/report/report_test.go`: verify totals, ratios, package grouping, and zero-production edge cases.
- Modify `internal/render/render.go`: render table, JSON, and Markdown from the expanded result.
- Create `internal/render/render_test.go`: verify stable output fragments for table and Markdown, and JSON field names.
- Modify `internal/cli/cli.go`: add output flags, include flags, ignore flags, and implement `report`, `packages`, `files`, and `tests`.
- Modify `internal/cli/cli_test.go`: cover CLI format, output file, default command, planned commands becoming real commands, and invalid flags.
- Modify `README.md`: document current commands, examples, and development workflow.

## Behavior Contract

- Default command: `gocensus .` is equivalent to `gocensus scan .`.
- Main command: `gocensus scan .`.
- Output formats: `table`, `json`, `markdown`.
- Report command: `gocensus report . --output census.md` writes Markdown.
- `.gitignore` is respected by default.
- `--no-gitignore` disables `.gitignore` filtering.
- `--exclude <pattern>` appends extra gitignore-style patterns.
- Production excludes tests, generated files, mocks, `vendor/`, `.git/`, `node_modules/`, hidden directories, and ignored paths by default.
- `--include-generated` includes generated files in production totals.
- `--include-mocks` includes mocks in production totals.
- Effective lines count unique source lines that contain non-comment Go tokens.
- Test declarations count top-level functions named `Test*`, `Benchmark*`, and `Example*` in `*_test.go` files.
- Every commit step runs `make check` first. Commit messages must not include attribution trailers.

## Task 1: Expand Public Result Model

**Files:**
- Modify: `census.go`
- Modify: `analyze_test.go`

- [ ] **Step 1: Write failing tests for the public API shape**

Replace `TestAnalyzeEmptyGoModule` in `analyze_test.go` with this expanded test and keep the existing imports:

```go
func TestAnalyzeEmptyGoModule(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := gocensus.Analyze(context.Background(), gocensus.Options{Root: dir})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.Root != dir {
		t.Fatalf("Root = %q, want %q", result.Root, dir)
	}
	if result.ModulePath != "example.com/app" {
		t.Fatalf("ModulePath = %q, want example.com/app", result.ModulePath)
	}
	if result.Files.Total != 0 {
		t.Fatalf("Files.Total = %d, want 0", result.Files.Total)
	}
	if len(result.Packages) != 0 {
		t.Fatalf("Packages length = %d, want 0", len(result.Packages))
	}
	if len(result.FileMetrics) != 0 {
		t.Fatalf("FileMetrics length = %d, want 0", len(result.FileMetrics))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./...
```

Expected: FAIL because `Result.Packages` and `Result.FileMetrics` do not exist.

- [ ] **Step 3: Expand public types**

In `census.go`, add these fields and types:

```go
type Options struct {
	Root             string
	NoGitignore      bool
	ExtraExcludes     []string
	IncludeGenerated bool
	IncludeMocks     bool
}

type Result struct {
	Root        string          `json:"root"`
	ModulePath  string          `json:"module_path"`
	Files       FileCounts      `json:"files"`
	Lines       LineCounts      `json:"lines"`
	Tests       TestCounts      `json:"tests"`
	Ratios      Ratios          `json:"ratios"`
	Packages    []PackageMetric `json:"packages"`
	FileMetrics []FileMetric    `json:"file_metrics"`
}

type Ratios struct {
	TestToProductionRaw       float64 `json:"test_to_production_raw"`
	TestToProductionEffective float64 `json:"test_to_production_effective"`
	TestShareEffective        float64 `json:"test_share_effective"`
	GeneratedShareRaw         float64 `json:"generated_share_raw"`
	MockShareRaw              float64 `json:"mock_share_raw"`
}

type PackageMetric struct {
	ImportPath string     `json:"import_path"`
	Dir        string     `json:"dir"`
	Files      FileCounts `json:"files"`
	Lines      LineCounts `json:"lines"`
	Tests      TestCounts `json:"tests"`
	Ratios     Ratios     `json:"ratios"`
}

type FileMetric struct {
	Path       string `json:"path"`
	Package   string `json:"package"`
	Kind      string `json:"kind"`
	Generated bool   `json:"generated"`
	RawLines  int    `json:"raw_lines"`
	CodeLines int    `json:"code_lines"`
	Tests     int    `json:"tests"`
	Benchmarks int   `json:"benchmarks"`
	Examples  int    `json:"examples"`
}
```

Keep existing `FileCounts`, `LineCounts`, `Metric`, and `TestCounts`.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Run lint and commit**

Run:

```bash
make check
git add census.go analyze_test.go
git commit -m "feat: expand census result model"
```

Expected: `make check` passes before the commit. Commit message has no trailers.

## Task 2: Discover Go Files With Excludes and .gitignore

**Files:**
- Create: `internal/discover/discover.go`
- Create: `internal/discover/discover_test.go`

- [ ] **Step 1: Keep discovery dependency-free**

Run:

```bash
go list -m all
```

Expected: only `github.com/arloliu/gocensus` is listed. Discovery uses the standard library so installing the CLI stays lightweight.

- [ ] **Step 2: Write failing discovery tests**

Create `internal/discover/discover_test.go`:

```go
package discover_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/arloliu/gocensus/internal/discover"
)

func TestGoFilesRespectsDefaultExcludesAndGitignore(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".gitignore", "ignored.go\nnested/ignored_dir/\n")
	write(t, root, "main.go", "package main\n")
	write(t, root, "ignored.go", "package main\n")
	write(t, root, "vendor/lib/lib.go", "package lib\n")
	write(t, root, ".hidden/hidden.go", "package hidden\n")
	write(t, root, "nested/keep.go", "package nested\n")
	write(t, root, "nested/ignored_dir/drop.go", "package ignored\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:         root,
		UseGitignore: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"main.go", "nested/keep.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func TestGoFilesCanDisableGitignore(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".gitignore", "ignored.go\n")
	write(t, root, "main.go", "package main\n")
	write(t, root, "ignored.go", "package main\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:         root,
		UseGitignore: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"ignored.go", "main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func TestGoFilesUsesExtraExcludes(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	write(t, root, "internal/drop.go", "package internal\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:          root,
		UseGitignore:  true,
		ExtraExcludes: []string{"internal/**"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func write(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func rels(t *testing.T, root string, files []string) []string {
	t.Helper()
	out := make([]string, 0, len(files))
	for _, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	slices.Sort(out)
	return out
}
```

- [ ] **Step 3: Run tests to verify failure**

Run:

```bash
go test ./internal/discover
```

Expected: FAIL because `internal/discover` does not exist.

- [ ] **Step 4: Implement discovery**

Create `internal/discover/discover.go`:

```go
package discover

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type Options struct {
	Root          string
	UseGitignore  bool
	ExtraExcludes []string
}

type matcher struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	domain  string
	pattern string
	dirOnly bool
}

func GoFiles(ctx context.Context, opts Options) ([]string, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	matcher, err := buildMatcher(root, opts)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if entry.IsDir() && isHardExcludedDir(entry.Name()) {
			return filepath.SkipDir
		}
		if matcher.match(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}

func buildMatcher(root string, opts Options) (matcher, error) {
	var patterns []ignorePattern
	if opts.UseGitignore {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() && path != root && isHardExcludedDir(entry.Name()) {
				return filepath.SkipDir
			}
			if entry.IsDir() || entry.Name() != ".gitignore" {
				return nil
			}
			relDir, err := filepath.Rel(root, filepath.Dir(path))
			if err != nil {
				return err
			}
			loaded, err := readPatterns(path, cleanDomain(relDir))
			if err != nil {
				return err
			}
			patterns = append(patterns, loaded...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for _, pattern := range opts.ExtraExcludes {
		if parsed, ok := parsePattern(pattern, ""); ok {
			patterns = append(patterns, parsed)
		}
	}
	return matcher{patterns: patterns}, nil
}

func readPatterns(path string, domain string) ([]ignorePattern, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var patterns []ignorePattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if parsed, ok := parsePattern(scanner.Text(), domain); ok {
			patterns = append(patterns, parsed)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

func parsePattern(raw string, domain string) (ignorePattern, bool) {
	pattern := strings.TrimSpace(raw)
	if pattern == "" || strings.HasPrefix(pattern, "#") || strings.HasPrefix(pattern, "!") {
		return ignorePattern{}, false
	}
	dirOnly := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimPrefix(strings.TrimSuffix(pattern, "/"), "/")
	if pattern == "" {
		return ignorePattern{}, false
	}
	return ignorePattern{
		domain:  domain,
		pattern: filepath.ToSlash(pattern),
		dirOnly: dirOnly,
	}, true
}

func (m matcher) match(rel string, isDir bool) bool {
	for _, pattern := range m.patterns {
		if pattern.match(rel, isDir) {
			return true
		}
	}
	return false
}

func (p ignorePattern) match(rel string, isDir bool) bool {
	target, ok := p.target(rel)
	if !ok {
		return false
	}
	if p.dirOnly {
		return isDir && target == p.pattern || strings.HasPrefix(target, p.pattern+"/")
	}
	if !strings.Contains(p.pattern, "/") {
		return path.Base(target) == p.pattern
	}
	matched, err := path.Match(p.pattern, target)
	if err != nil {
		return false
	}
	return matched
}

func (p ignorePattern) target(rel string) (string, bool) {
	if p.domain == "" {
		return rel, true
	}
	if rel == p.domain {
		return "", true
	}
	prefix := p.domain + "/"
	if !strings.HasPrefix(rel, prefix) {
		return "", false
	}
	return strings.TrimPrefix(rel, prefix), true
}

func cleanDomain(relDir string) string {
	relDir = filepath.ToSlash(relDir)
	if relDir == "." {
		return ""
	}
	return relDir
}

func isHardExcludedDir(name string) bool {
	if name == ".git" || name == "vendor" || name == "node_modules" {
		return true
	}
	return strings.HasPrefix(name, ".")
}
```

- [ ] **Step 5: Run tests to verify pass**

Run:

```bash
go test ./internal/discover
```

Expected: PASS.

- [ ] **Step 6: Run full checks and commit**

Run:

```bash
make check
	git add internal/discover docs/plans/gocensus-scan.md
git commit -m "feat: discover go files with gitignore"
```

Expected: `make check` passes before the commit.

## Task 3: Classify Files

**Files:**
- Create: `internal/classify/classify.go`
- Create: `internal/classify/classify_test.go`

- [ ] **Step 1: Write failing classification tests**

Create `internal/classify/classify_test.go`:

```go
package classify_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus/internal/classify"
)

func TestFileKind(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		rel     string
		content string
		want    classify.Kind
	}{
		{rel: "main.go", content: "package main\n", want: classify.KindProduction},
		{rel: "main_test.go", content: "package main\n", want: classify.KindTest},
		{rel: "mock_client.go", content: "package main\n", want: classify.KindMock},
		{rel: "client_mock.go", content: "package main\n", want: classify.KindMock},
		{rel: "service.pb.go", content: "package main\n", want: classify.KindGenerated},
		{rel: "generated.go", content: "// Code generated by tool; DO NOT EDIT.\npackage main\n", want: classify.KindGenerated},
	}

	for _, tc := range cases {
		t.Run(tc.rel, func(t *testing.T) {
			path := filepath.Join(root, filepath.FromSlash(tc.rel))
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := classify.File(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("File(%q) = %q, want %q", tc.rel, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/classify
```

Expected: FAIL because `internal/classify` does not exist.

- [ ] **Step 3: Implement classification**

Create `internal/classify/classify.go`:

```go
package classify

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Kind string

const (
	KindProduction Kind = "production"
	KindTest       Kind = "test"
	KindGenerated  Kind = "generated"
	KindMock       Kind = "mock"
)

func File(path string) (Kind, error) {
	base := filepath.Base(path)
	lower := strings.ToLower(base)

	if strings.HasSuffix(lower, ".pb.go") || hasGeneratedMarker(path) {
		return KindGenerated, nil
	}
	if strings.HasSuffix(lower, "_test.go") {
		return KindTest, nil
	}
	if strings.Contains(lower, "mock") {
		return KindMock, nil
	}
	return KindProduction, nil
}

func hasGeneratedMarker(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for i := 0; i < 20 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") && strings.Contains(line, "DO NOT EDIT.") {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```bash
go test ./internal/classify
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add internal/classify
git commit -m "feat: classify go source files"
```

Expected: `make check` passes before the commit.

## Task 4: Count Lines and Test Declarations

**Files:**
- Create: `internal/count/count.go`
- Create: `internal/count/count_test.go`

- [ ] **Step 1: Write failing counting tests**

Create `internal/count/count_test.go`:

```go
package count_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus/internal/count"
)

func TestFileCountsRawAndEffectiveLines(t *testing.T) {
	path := write(t, `package main

// package comment
func main() { // inline comment
	println("hi")
}

/*
block comment
*/
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.RawLines != 10 {
		t.Fatalf("RawLines = %d, want 10", metrics.RawLines)
	}
	if metrics.CodeLines != 4 {
		t.Fatalf("CodeLines = %d, want 4", metrics.CodeLines)
	}
}

func TestFileCountsTestDeclarations(t *testing.T) {
	path := write(t, `package main

import "testing"

func TestOne(t *testing.T) {}
func Testhelper(t *testing.T) {}
func BenchmarkOne(b *testing.B) {}
func ExampleOne() {}
func helper() {}
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.Tests != 2 {
		t.Fatalf("Tests = %d, want 2", metrics.Tests)
	}
	if metrics.Benchmarks != 1 {
		t.Fatalf("Benchmarks = %d, want 1", metrics.Benchmarks)
	}
	if metrics.Examples != 1 {
		t.Fatalf("Examples = %d, want 1", metrics.Examples)
	}
}

func write(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "file_test.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/count
```

Expected: FAIL because `internal/count` does not exist.

- [ ] **Step 3: Implement counting**

Create `internal/count/count.go`:

```go
package count

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"strings"
)

type Metrics struct {
	RawLines   int
	CodeLines  int
	Tests      int
	Benchmarks int
	Examples   int
	Package    string
}

func File(path string) (Metrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Metrics{}, err
	}

	metrics := Metrics{
		RawLines:  rawLines(content),
		CodeLines: effectiveLines(content),
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		return metrics, nil
	}
	metrics.Package = file.Name.Name
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		switch {
		case strings.HasPrefix(fn.Name.Name, "Test"):
			metrics.Tests++
		case strings.HasPrefix(fn.Name.Name, "Benchmark"):
			metrics.Benchmarks++
		case strings.HasPrefix(fn.Name.Name, "Example"):
			metrics.Examples++
		}
	}
	return metrics, nil
}

func rawLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func effectiveLines(content []byte) int {
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(content))
	var scan scanner.Scanner
	scan.Init(file, content, nil, 0)

	lines := map[int]struct{}{}
	for {
		pos, tok, _ := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			continue
		}
		lines[fileSet.Position(pos).Line] = struct{}{}
	}
	return len(lines)
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```bash
go test ./internal/count
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add internal/count
git commit -m "feat: count go source metrics"
```

Expected: `make check` passes before the commit.

## Task 5: Aggregate Repository and Package Metrics

**Files:**
- Create: `internal/report/report.go`
- Create: `internal/report/report_test.go`

- [ ] **Step 1: Write failing report tests**

Create `internal/report/report_test.go`:

```go
package report_test

import (
	"reflect"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/report"
)

func TestBuildAggregatesRepoAndPackageMetrics(t *testing.T) {
	files := []gocensus.FileMetric{
		{Path: "main.go", Package: "main", Kind: "production", RawLines: 10, CodeLines: 8},
		{Path: "main_test.go", Package: "main", Kind: "test", RawLines: 6, CodeLines: 4, Tests: 2},
		{Path: "mock_client.go", Package: "main", Kind: "mock", RawLines: 3, CodeLines: 2},
		{Path: "service.pb.go", Package: "main", Kind: "generated", RawLines: 20, CodeLines: 18},
	}

	result := report.Build(report.Input{
		Root:        "/repo",
		ModulePath:  "example.com/app",
		FileMetrics: files,
	})

	if result.Files.Total != 4 {
		t.Fatalf("Files.Total = %d, want 4", result.Files.Total)
	}
	if result.Files.Production != 1 || result.Files.Tests != 1 || result.Files.Mocks != 1 || result.Files.Generated != 1 {
		t.Fatalf("file buckets = %#v", result.Files)
	}
	if result.Lines.Production.Raw != 10 || result.Lines.Tests.Raw != 6 {
		t.Fatalf("line buckets = %#v", result.Lines)
	}
	if result.Tests.Tests != 2 {
		t.Fatalf("Tests.Tests = %d, want 2", result.Tests.Tests)
	}
	if result.Ratios.TestToProductionEffective != 0.5 {
		t.Fatalf("effective ratio = %v, want 0.5", result.Ratios.TestToProductionEffective)
	}
	if len(result.Packages) != 1 {
		t.Fatalf("package count = %d, want 1", len(result.Packages))
	}
	if !reflect.DeepEqual(result.FileMetrics, files) {
		t.Fatalf("FileMetrics changed")
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/report
```

Expected: FAIL because `internal/report` does not exist.

- [ ] **Step 3: Implement aggregation**

Create `internal/report/report.go` with deterministic package sorting and ratio helpers:

```go
package report

import (
	"slices"

	"github.com/arloliu/gocensus"
)

type Input struct {
	Root        string
	ModulePath  string
	FileMetrics []gocensus.FileMetric
}

func Build(input Input) gocensus.Result {
	result := gocensus.Result{
		Root:        input.Root,
		ModulePath:  input.ModulePath,
		FileMetrics: slices.Clone(input.FileMetrics),
	}

	byPackage := map[string]*gocensus.PackageMetric{}
	for _, file := range input.FileMetrics {
		addFile(&result.Files, &result.Lines, &result.Tests, file)
		pkg := byPackage[file.Package]
		if pkg == nil {
			pkg = &gocensus.PackageMetric{
				ImportPath: file.Package,
				Dir:        file.Package,
			}
			byPackage[file.Package] = pkg
		}
		addFile(&pkg.Files, &pkg.Lines, &pkg.Tests, file)
	}
	result.Ratios = ratios(result.Lines)

	for _, pkg := range byPackage {
		pkg.Ratios = ratios(pkg.Lines)
		result.Packages = append(result.Packages, *pkg)
	}
	slices.SortFunc(result.Packages, func(a, b gocensus.PackageMetric) int {
		if a.ImportPath < b.ImportPath {
			return -1
		}
		if a.ImportPath > b.ImportPath {
			return 1
		}
		return 0
	})
	return result
}

func addFile(files *gocensus.FileCounts, lines *gocensus.LineCounts, tests *gocensus.TestCounts, file gocensus.FileMetric) {
	files.Total++
	tests.Tests += file.Tests
	tests.Benchmarks += file.Benchmarks
	tests.Examples += file.Examples

	metric := gocensus.Metric{Raw: file.RawLines, Effective: file.CodeLines}
	switch file.Kind {
	case "production":
		files.Production++
		lines.Production.Raw += metric.Raw
		lines.Production.Effective += metric.Effective
	case "test":
		files.Tests++
		lines.Tests.Raw += metric.Raw
		lines.Tests.Effective += metric.Effective
	case "generated":
		files.Generated++
		lines.Generated.Raw += metric.Raw
		lines.Generated.Effective += metric.Effective
	case "mock":
		files.Mocks++
		lines.Mocks.Raw += metric.Raw
		lines.Mocks.Effective += metric.Effective
	}
}

func ratios(lines gocensus.LineCounts) gocensus.Ratios {
	totalRaw := lines.Production.Raw + lines.Tests.Raw + lines.Generated.Raw + lines.Mocks.Raw
	totalEffective := lines.Production.Effective + lines.Tests.Effective
	return gocensus.Ratios{
		TestToProductionRaw:       divide(lines.Tests.Raw, lines.Production.Raw),
		TestToProductionEffective: divide(lines.Tests.Effective, lines.Production.Effective),
		TestShareEffective:        divide(lines.Tests.Effective, totalEffective),
		GeneratedShareRaw:         divide(lines.Generated.Raw, totalRaw),
		MockShareRaw:              divide(lines.Mocks.Raw, totalRaw),
	}
}

func divide(numerator int, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```bash
go test ./internal/report
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add internal/report
git commit -m "feat: aggregate census metrics"
```

Expected: `make check` passes before the commit.

## Task 6: Wire Analyze End-to-End

**Files:**
- Modify: `census.go`
- Modify: `analyze_test.go`

- [ ] **Step 1: Write failing end-to-end analysis test**

Add this test to `analyze_test.go`:

```go
func TestAnalyzeCountsRepositoryMetrics(t *testing.T) {
	dir := t.TempDir()
writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, ".gitignore", "ignored.go\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n")
	writeFile(t, dir, "mock_client.go", "package main\n\ntype mockClient struct{}\n")
	writeFile(t, dir, "service.pb.go", "// Code generated by protoc; DO NOT EDIT.\npackage main\n")
	writeFile(t, dir, "ignored.go", "package main\n")

	result, err := gocensus.Analyze(context.Background(), gocensus.Options{
		Root: dir,
	})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.Files.Total != 4 {
		t.Fatalf("Files.Total = %d, want 4", result.Files.Total)
	}
	if result.Files.Production != 1 || result.Files.Tests != 1 || result.Files.Mocks != 1 || result.Files.Generated != 1 {
		t.Fatalf("file buckets = %#v", result.Files)
	}
	if result.Tests.Tests != 1 {
		t.Fatalf("Tests.Tests = %d, want 1", result.Tests.Tests)
	}
	if len(result.FileMetrics) != 4 {
		t.Fatalf("FileMetrics length = %d, want 4", len(result.FileMetrics))
	}
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
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./...
```

Expected: FAIL because `Analyze` still returns only root and module metadata.

- [ ] **Step 3: Implement Analyze orchestration**

Modify `census.go` imports to include:

```go
	"github.com/arloliu/gocensus/internal/classify"
	"github.com/arloliu/gocensus/internal/count"
	"github.com/arloliu/gocensus/internal/discover"
	"github.com/arloliu/gocensus/internal/report"
```

Update `Analyze` after module path lookup:

```go
	files, err := discover.GoFiles(ctx, discover.Options{
		Root:          absRoot,
		UseGitignore:  !opts.NoGitignore,
		ExtraExcludes: opts.ExtraExcludes,
	})
	if err != nil {
		return Result{}, err
	}

	fileMetrics := make([]FileMetric, 0, len(files))
	for _, path := range files {
		kind, err := classify.File(path)
		if err != nil {
			return Result{}, err
		}
		metrics, err := count.File(path)
		if err != nil {
			return Result{}, err
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return Result{}, err
		}
		kindString := string(kind)
		if kind == classify.KindGenerated && opts.IncludeGenerated {
			kindString = string(classify.KindProduction)
		}
		if kind == classify.KindMock && opts.IncludeMocks {
			kindString = string(classify.KindProduction)
		}
		fileMetrics = append(fileMetrics, FileMetric{
			Path:       filepath.ToSlash(rel),
			Package:   metrics.Package,
			Kind:      kindString,
			Generated: kind == classify.KindGenerated,
			RawLines:  metrics.RawLines,
			CodeLines: metrics.CodeLines,
			Tests:     metrics.Tests,
			Benchmarks: metrics.Benchmarks,
			Examples:  metrics.Examples,
		})
	}

	return report.Build(report.Input{
		Root:        absRoot,
		ModulePath:  modulePath,
		FileMetrics: fileMetrics,
	}), nil
```

Then remove the old direct `Result{Root: absRoot, ModulePath: modulePath}` return.

- [ ] **Step 4: Run tests to verify pass**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add census.go analyze_test.go
git commit -m "feat: analyze repositories end to end"
```

Expected: `make check` passes before the commit.

## Task 7: Render Table, JSON, and Markdown Reports

**Files:**
- Modify: `internal/render/render.go`
- Create: `internal/render/render_test.go`

- [ ] **Step 1: Write failing renderer tests**

Create `internal/render/render_test.go`:

```go
package render_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/render"
)

func TestTableIncludesCoreSections(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "table")
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"Go Census: example.com/app", "Files", "Lines", "Ratios", "Tests"} {
		if !strings.Contains(text, want) {
			t.Fatalf("table missing %q:\n%s", want, text)
		}
	}
}

func TestMarkdownIncludesPackageTable(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "markdown")
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "| Package | Prod Lines | Test Lines | Test Ratio |") {
		t.Fatalf("markdown missing package table:\n%s", text)
	}
}

func TestJSONUsesStableFieldNames(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "json")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["module_path"]; !ok {
		t.Fatalf("json missing module_path: %s", out.String())
	}
	if _, ok := payload["file_metrics"]; !ok {
		t.Fatalf("json missing file_metrics: %s", out.String())
	}
}

func sample() gocensus.Result {
	return gocensus.Result{
		Root:       "/repo",
		ModulePath: "example.com/app",
		Files:      gocensus.FileCounts{Total: 2, Production: 1, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Raw: 10, Effective: 8},
			Tests:      gocensus.Metric{Raw: 6, Effective: 4},
		},
		Tests:  gocensus.TestCounts{Tests: 2},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
		Packages: []gocensus.PackageMetric{{
			ImportPath: "main",
			Lines: gocensus.LineCounts{
				Production: gocensus.Metric{Raw: 10, Effective: 8},
				Tests:      gocensus.Metric{Raw: 6, Effective: 4},
			},
			Ratios: gocensus.Ratios{TestToProductionEffective: 0.5},
		}},
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/render
```

Expected: FAIL because table and Markdown are still minimal.

- [ ] **Step 3: Implement stable renderers**

Modify `internal/render/render.go` so table output includes `Files`, `Lines`, `Ratios`, and `Tests`, Markdown includes summary and package tables, and JSON keeps using `json.Encoder` with indentation.

Use this formatting helper:

```go
func pct(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func ratio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
```

For package Markdown rows, render:

```go
_, err := fmt.Fprintf(w, "| %s | %d | %d | %s |\n",
	pkg.ImportPath,
	pkg.Lines.Production.Effective,
	pkg.Lines.Tests.Effective,
	ratio(pkg.Ratios.TestToProductionEffective),
)
```

- [ ] **Step 4: Run renderer tests**

Run:

```bash
go test ./internal/render
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add internal/render
git commit -m "feat: render census reports"
```

Expected: `make check` passes before the commit.

## Task 8: Complete CLI Commands and Flags

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`

- [ ] **Step 1: Write failing CLI tests**

Add tests to `internal/cli/cli_test.go`:

```go
func TestRunReportWritesMarkdownFile(t *testing.T) {
	dir := writeModule(t)
	output := filepath.Join(t.TempDir(), "census.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"report", dir, "--output", output}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "# Go Census: example.com/app") {
		t.Fatalf("report content = %q", string(content))
	}
}

func TestRunPackagesPrintsPackageView(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"packages", dir}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Packages") {
		t.Fatalf("stdout = %q, want package view", stdout.String())
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./internal/cli
```

Expected: FAIL because planned commands still exit with code 2.

- [ ] **Step 3: Implement command parsing**

In `internal/cli/cli.go`, replace planned-command handling with real command handlers:

- `scan`: render full result using selected format.
- `report`: default `--format markdown`, require or accept `--output`.
- `packages`: render package-focused table.
- `files`: render file-focused table with `--top`.
- `tests`: render test inventory table.

Use one shared parser struct:

```go
type commonArgs struct {
	root             string
	format           string
	output           string
	top              int
	sort             string
	useGitignore     bool
	extraExcludes     []string
	includeGenerated bool
	includeMocks     bool
}
```

Default values:

```go
commonArgs{
	root:         ".",
	format:       "table",
	useGitignore: true,
	top:          20,
	sort:         "path",
}
```

Support these flags in any position:

```text
--format <table|json|markdown>
--format=<table|json|markdown>
--output <path>
--output=<path>
--top <n>
--top=<n>
--sort <path|prod-lines|test-lines|test-ratio>
--sort=<path|prod-lines|test-lines|test-ratio>
--no-gitignore
--exclude <pattern>
--exclude=<pattern>
--include-generated
--include-mocks
```

- [ ] **Step 4: Run CLI tests**

Run:

```bash
go test ./internal/cli
```

Expected: PASS.

- [ ] **Step 5: Run full checks and commit**

Run:

```bash
make check
git add internal/cli
git commit -m "feat: implement gocensus cli commands"
```

Expected: `make check` passes before the commit.

## Task 9: Update README and Final Verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README with final command examples**

Replace the Usage section in `README.md` with:

````markdown
## Usage

```bash
gocensus scan .
gocensus scan . --format json
gocensus scan . --format markdown
gocensus report . --output reports/census.md
gocensus packages . --sort test-ratio
gocensus files . --top 20
gocensus tests .
```

The default command is `scan`, so this is equivalent:

```bash
gocensus .
```

### Ignore and Bucket Options

```bash
gocensus scan . --no-gitignore
gocensus scan . --exclude 'internal/generated/**'
gocensus scan . --include-generated
gocensus scan . --include-mocks
```
````

- [ ] **Step 2: Run verification commands**

Run:

```bash
make check
./bin/gocensus scan . --format table
./bin/gocensus scan . --format json
./bin/gocensus report . --output reports/census.md
test -s reports/census.md
```

Expected:

- `make check` passes.
- `scan --format table` prints `Go Census: github.com/arloliu/gocensus`.
- `scan --format json` prints valid JSON containing `"module_path": "github.com/arloliu/gocensus"`.
- `reports/census.md` exists and is non-empty.

- [ ] **Step 3: Commit documentation**

Run:

```bash
git add README.md reports/.gitkeep
git commit -m "docs: document gocensus usage"
```

Expected: Commit succeeds with no trailers.

## Self-Review Checklist

- Spec coverage: The plan covers library API, CLI `scan`, default command, report output, JSON/Markdown/table renderers, `.gitignore`, production/test/generated/mock classification, effective line counting, package metrics, file metrics, and test declarations.
- Placeholder scan: The plan contains concrete file paths, commands, test code, expected failures, expected passes, and commit messages.
- Type consistency: Public types are defined in Task 1 and reused by reporting, rendering, CLI, and end-to-end analysis tasks.
- Risk notes: Task 2 implements a focused `.gitignore` matcher for common ignore patterns instead of adding a large dependency. Task 6 contains the main integration risk because it joins discovery, classification, counting, and aggregation.
