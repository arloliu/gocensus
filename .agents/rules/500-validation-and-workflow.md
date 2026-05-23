# 500 - Validation and Workflow

Apply before validation, commits, or completion claims.

## Normal Gates

After Go changes:

```bash
gofmt -w <changed-go-files>
go test ./...
golangci-lint run ./...
go build ./cmd/gocensus
```

Preferred full gate:

```bash
make check
```

## Make Targets

```bash
make fmt
make lint
make test
make build
make check
```

## Before Commit

- Always run linter and fix linting issues before commit.
- Prefer `make check` before commit.
- Verify README/docs changed when exported API or CLI behavior changed.
- Do not commit generated binaries from `bin/`.

