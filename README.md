# gocensus

[![CI](https://github.com/arloliu/gocensus/actions/workflows/ci.yml/badge.svg)](https://github.com/arloliu/gocensus/actions/workflows/ci.yml)

`gocensus` is a Go CLI and library that shows how much production code, test code, generated code, and mock code a repository has.

It reports line counts, ratios, package/file breakdowns, and test inventory so you can quickly understand the shape of a Go codebase.

## What It Tells You

- How many Go files are production, tests, generated code, or mocks.
- Raw line counts and effective code line counts.
- Test-to-production ratios and test share.
- Package-level and file-level hot spots.
- Top-level tests, statically countable subtests, dynamic subtest sites, benchmarks, and examples.
- Markdown and JSON reports for docs or CI artifacts.

## Install

Requires Go 1.21 or newer.

```bash
go install github.com/arloliu/gocensus/cmd/gocensus@latest
```

Install a specific release:

```bash
go install github.com/arloliu/gocensus/cmd/gocensus@v0.1.0
```

Check the installed version:

```bash
gocensus version
```

## Quick Start

Run the main census for the current repository:

```bash
gocensus scan .
```

`scan` is the default command, so this is equivalent:

```bash
gocensus .
```

Example output:

```text
Go Census: github.com/arloliu/gocensus
Scope: production excludes generated and mock files

Overview
  Go files: 24    Packages: 13    Known test cases: 46

Code Mix
  Kind                 Files   Raw Lines   Effective Lines
  Production Scope       14       1,538             1,320
  Tests                  10       1,103               848
  Excluded Generated      0           0                 0
  Excluded Mocks          0           0                 0
  Total                  24       2,641             2,168

Ratios
  Test / Production Scope       0.64:1
  Test Share                     39.1%
```

## Commands

| Command | Purpose |
| --- | --- |
| `gocensus scan [root]` | Print the main repository census. |
| `gocensus report [root]` | Write a repository report, Markdown by default. |
| `gocensus packages [root]` | Show package-level production and test metrics. |
| `gocensus files [root]` | Show file-level classification and line counts. |
| `gocensus tests [root]` | Summarize tests, subtests, benchmarks, and examples. |
| `gocensus who [root]` | Rank contributors from Git history. |
| `gocensus diff [root]` | Compare scan metrics between two Git refs. |
| `gocensus hotspots [root]` | Rank human-authored Go file hotspots by size and Git churn. |
| `gocensus version` | Print the CLI version. |

Use command-specific help to see flags:

```bash
gocensus --help
gocensus scan --help
gocensus help packages
```

## Common Usage

Print the default table output:

```bash
gocensus scan .
```

Print JSON:

```bash
gocensus scan . -f json
```

Print Markdown:

```bash
gocensus scan . -f markdown
```

Write a Markdown report:

```bash
gocensus report . -o census.md
```

Sort packages by test ratio:

```bash
gocensus packages . -s test-ratio
```

Show the largest files:

```bash
gocensus files . -n 20
```

Show all files:

```bash
gocensus files . -n 0
```

Show test inventory:

```bash
gocensus tests .
```

Rank contributors by commit count:

```bash
gocensus who .
```

Rank contributors in the recommended human-authored Go scope:

```bash
gocensus who . --go-only
```

Rank contributors by removed lines:

```bash
gocensus who . --go-only --by removed
```

Compare scan metrics between two refs:

```bash
gocensus diff . --base v0.1.0 --head HEAD
gocensus diff . --base main --head feature/my-change -f markdown -o diff.md
```

Rank file hotspots by current size plus Git churn:

```bash
gocensus hotspots . --since 90.days
gocensus hotspots . --by churn -n 20
```

## Output Formats

`scan`, `report`, `diff`, `hotspots`, and `who` support:

- `table`
- `json`
- `markdown`

Examples:

```bash
gocensus scan . -f json
gocensus scan . -f markdown
gocensus report . -f markdown -o census.md
```

Human-readable terminal output uses color automatically when the terminal advertises color support. Use `--color always` to force RGB color, `--color never` or `--no-color` to disable it, and `--color auto` to keep the default detection behavior. JSON, Markdown, and files written with `--output` stay plain.

## Ignore and Bucket Options

`gocensus` reads `.gitignore` by default.

```bash
gocensus scan . --no-gitignore
gocensus scan . -x 'internal/generated/**'
gocensus scan . --include-generated
gocensus scan . --include-mocks
```

Global analysis flags:

| Flag | Meaning |
| --- | --- |
| `--no-gitignore` | Do not read `.gitignore` exclude rules. |
| `-x, --exclude PATTERN` | Exclude paths matching a pattern. Can be repeated. |
| `--include-generated` | Count generated files as production code for scan/report views and include generated paths in contributor rankings. |
| `--include-mocks` | Count mock files as production code for scan/report views and include mock paths in contributor rankings. |

## Metric Meanings

| Field | Meaning |
| --- | --- |
| Raw Lines | Physical lines, including blanks and comments. |
| Effective Lines | Lines containing non-comment Go tokens. |
| Production Scope | Non-test Go files counted as production; the scope line says whether generated and mock files are included. |
| Tests | `*_test.go` files. |
| Test / Production Scope | Effective test lines divided by effective production-scope lines. |
| Test Share | Effective test lines divided by production plus test effective lines. |
| Known Test Cases | Top-level tests plus statically countable subtests. |
| Dynamic Subtest Sites | `t.Run` or `b.Run` call sites where runtime data controls the case count. |
| Hotspot Score | Effective production lines plus Git churn. |
| Git Churn | Added plus removed lines from `git log --numstat`. |
| Package Test Ratio | Package test effective lines divided by package production effective lines. |
| Diff Delta | Head value minus base value. Positive values are shown with `+` in table output. |

## Test Inventory

```text
Tests
  Known Test Cases         12
  Top-level Tests           5
  Static Subtests           7
  Dynamic Subtest Sites     2

Benchmarks
  Known Benchmark Cases     3
  Top-level Benchmarks      1
  Static Subbenchmarks      2
  Dynamic Benchmark Sites   1

Examples
  Examples                  4
```

`Known Test Cases` is the sum of top-level `TestXxx` functions and subtests whose case count can be determined statically.

Dynamic subtest sites are reported separately because their runtime case count is not knowable from source alone.

## Contributor Rankings

`gocensus who` reads Git history for tracked files under the requested root. It combines factual Git diffstat metrics with transparent commit-message heuristics.

Like `scan`, contributor rankings exclude generated and mock paths by default. Add those paths back explicitly:

```bash
gocensus who . --include-generated
gocensus who . --include-mocks
```

Use `--go-only` to limit the ranking to Go paths while keeping the same generated/mock defaults:

```bash
gocensus who . --go-only
gocensus who . --go-only --include-generated
gocensus who . --go-only --include-mocks
```

Contribution path filtering is path-based so it works across historical commits, including files that were later deleted. It does not inspect old file contents for generated-code comments.

Ranking choices:

| Sort | Meaning |
| --- | --- |
| `commits` | Number of commits by author. |
| `features` | Commits whose subject looks like feature work, such as `feat:` or `add`. |
| `fixes` | Commits whose subject looks like a bug or issue fix, such as `fix:` or `closes #123`. |
| `refactors` | Commits whose subject looks like refactoring work, such as `refactor:`. |
| `added` | Lines added from Git numstat. |
| `removed` | Lines removed from Git numstat. |
| `net` | Added lines minus removed lines. |
| `shrink` | Largest net reduction, computed as removed lines minus added lines. |
| `churn` | Added plus removed lines. |
| `files` | Unique file paths touched. |
| `active-days` | Distinct commit dates by author. |

Examples:

```bash
gocensus who . --go-only
gocensus who . --by churn -n 20
gocensus who . --go-only --by churn -n 20
gocensus who . --since 2026-01-01 --until 2026-03-31
gocensus who . -f markdown -o contributors.md
```

Feature, fix, and refactor counts are message-classified intent metrics, not semantic proof of what changed. Line, file, commit, and active-day counts come from Git history.

## Library Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arloliu/gocensus"
)

func main() {
	result, err := gocensus.Analyze(context.Background(), gocensus.Options{
		Root: ".",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.ModulePath)
	fmt.Println(result.Lines.Production.Effective)
}
```

## Development

```bash
make fmt
make lint
make test
make build
make check
```

Build a binary with an explicit version:

```bash
make build VERSION=v0.1.0
./bin/gocensus version
```

By default, `make build` uses `git describe --tags --always --dirty` as the version and injects it with `-ldflags`.

When installing from source with Go, use a tagged module version:

```bash
go install github.com/arloliu/gocensus/cmd/gocensus@v0.1.0
gocensus version
```

For `go install ...@v0.1.0`, `gocensus version` uses Go build metadata embedded by the toolchain. For local development builds without a module version, it falls back to `dev` unless the Makefile injects a git tag or hash.

## License

MIT. See [LICENSE](LICENSE).
