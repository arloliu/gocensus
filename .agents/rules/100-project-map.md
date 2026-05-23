# 100 - Project Map

## Identity

- **Project:** gocensus
- **Module:** `github.com/arloliu/gocensus`
- **Purpose:** Go repository census tool for files, lines, tests, ratios, and reports.
- **Compatibility floor:** Go 1.21+.
- **Primary command:** `gocensus scan .`

## Current Layout

```text
gocensus/
├── census.go                 # Public library API
├── cmd/gocensus/             # CLI entrypoint
├── internal/cli/             # CLI parsing and command routing
├── internal/render/          # Table, JSON, Markdown renderers
├── docs/plans/               # Reviewable implementation plans
├── docs/specs/               # Design/spec artifacts
└── reports/                  # Generated report artifacts
```

Planned internal packages:

```text
internal/discover/            # Repo walking, .gitignore, excludes
internal/classify/            # production/test/generated/mock classification
internal/count/               # raw/effective lines, test declarations
internal/report/              # aggregation and ratios
```

## Dependency Policy

- Prefer the standard library.
- Keep dependencies small because this is a CLI users install from source.
- Before adding a dependency, check its `go` directive so it does not raise the Go 1.21 floor.
- If a dependency is only needed for convenience and has a high minimum Go version, avoid it.
