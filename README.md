# gocensus

`gocensus` is a Go repository census tool for files, lines, tests, and reports.

## Usage

```bash
gocensus scan .
gocensus scan . --format json
gocensus scan . --format markdown
gocensus report . --output census.md
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

## Development

Requires Go 1.21 or newer.

```bash
make fmt
make lint
make test
make build
```

Run `make check` before committing.
