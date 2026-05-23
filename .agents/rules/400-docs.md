# 400 - Docs

Apply when editing exported API, README, examples, or plan artifacts.

## Public API

- Every exported symbol must have a Go doc comment.
- The first sentence starts with the symbol name.
- Keep comments concise unless behavior has edge cases.

## README

Keep README examples current with the CLI:

```bash
gocensus scan .
gocensus scan . --format json
gocensus report . --output census.md
```

Mention the supported Go version when changing compatibility.

## Plans and Reports

- Store implementation plans in `docs/plans/`.
- Store design/spec artifacts in `docs/specs/`.
- Store generated sample reports in `reports/` only when they are intentional artifacts.
