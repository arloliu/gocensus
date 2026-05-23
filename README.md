# gocensus

`gocensus` is a Go repository census tool for files, lines, tests, and reports.

## Usage

```bash
gocensus scan .
gocensus scan . -f json
gocensus scan . -f markdown
gocensus report . -o census.md
gocensus packages . -s test-ratio
gocensus files . -n 20
gocensus tests .
gocensus version
```

The default command is `scan`, so this is equivalent:

```bash
gocensus .
```

Use `--help` at the root or on a subcommand to see command-specific flags:

```bash
gocensus --help
gocensus scan --help
gocensus tests --help
gocensus help packages
```

### Test Inventory

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

`Known Test Cases` is the sum of top-level `TestXxx` functions and subtests whose case count can be determined statically. Dynamic subtest sites are reported separately because their runtime case count is not knowable from source alone.

### Ignore and Bucket Options

```bash
gocensus scan . --no-gitignore
gocensus scan . -x 'internal/generated/**'
gocensus scan . --include-generated
gocensus scan . --include-mocks
```

## Development

Requires Go 1.21 or newer.

```bash
make fmt
make lint
make test
make build
```

Build a binary with an explicit version:

```bash
make build VERSION=v0.1.0
./bin/gocensus version
```

By default, `make build` uses `git describe --tags --always --dirty` as the version and injects it with `-ldflags`. Release builds can set `VERSION` explicitly.

When installing from source with Go, use a tagged module version:

```bash
go install github.com/arloliu/gocensus/cmd/gocensus@v0.1.0
gocensus version
```

For `go install ...@v0.1.0`, `gocensus version` uses Go build metadata embedded by the toolchain. For local development builds without a module version, it falls back to `dev` unless the Makefile injects a git tag or hash.

Run `make check` before committing.
