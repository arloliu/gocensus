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

Overview
  Go files: 24    Packages: 13    Known test cases: 46

Code Mix
  Kind          Files   Raw Lines   Effective Lines
  Production      14       1,538             1,320
  Tests           10       1,103               848
  Generated        0           0                 0
  Mocks            0           0                 0
  Total           24       2,641             2,168

Ratios
  Test / Production       0.64:1
  Test Share               39.1%
  Generated Share           0.0%
  Mock Share                0.0%
```

## Commands

| Command | Purpose |
| --- | --- |
| `gocensus scan [root]` | Print the main repository census. |
| `gocensus report [root]` | Write a repository report, Markdown by default. |
| `gocensus packages [root]` | Show package-level production and test metrics. |
| `gocensus files [root]` | Show file-level classification and line counts. |
| `gocensus tests [root]` | Summarize tests, subtests, benchmarks, and examples. |
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

## Output Formats

`scan` and `report` support:

- `table`
- `json`
- `markdown`

Examples:

```bash
gocensus scan . -f json
gocensus scan . -f markdown
gocensus report . -f markdown -o census.md
```

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
| `--include-generated` | Count generated files as production code. |
| `--include-mocks` | Count mock files as production code. |

## Metric Meanings

| Field | Meaning |
| --- | --- |
| Raw Lines | Physical lines, including blanks and comments. |
| Effective Lines | Lines containing non-comment Go tokens. |
| Production | Non-test Go files, excluding generated files and mocks by default. |
| Tests | `*_test.go` files. |
| Test / Production | Effective test lines divided by effective production lines. |
| Test Share | Effective test lines divided by production plus test effective lines. |
| Generated Share | Generated raw lines divided by total raw lines. |
| Mock Share | Mock raw lines divided by total raw lines. |
| Known Test Cases | Top-level tests plus statically countable subtests. |
| Dynamic Subtest Sites | `t.Run` or `b.Run` call sites where runtime data controls the case count. |

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
