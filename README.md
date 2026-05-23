# gocensus

`gocensus` is a Go repository census tool for files, lines, tests, and reports.

## Usage

```bash
gocensus scan .
gocensus scan . --format json
gocensus report . --output census.md
gocensus packages . --sort test-ratio
gocensus files . --top 20
gocensus tests .
```

The default command is `scan`, so this is equivalent:

```bash
gocensus .
```

## Development

Requires Go 1.21 or newer.

```bash
make fmt
make lint
make test
make build
```

Run `make check` before committing.
