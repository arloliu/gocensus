# gocensus Agent Rules Index

Read `000-agent-contract.md` for every task, then load only the rules that match the work.

## Default Load

- Go implementation: `000`, `100`, `200`, `300`, `500`.
- Docs or exported API: add `400`.
- Commits, branches, or PR text: add `550`.
- For tiny docs-only edits: `000` and `400` are enough.

## Rules

- [000-agent-contract.md](000-agent-contract.md) - evidence, scope, verification.
- [100-project-map.md](100-project-map.md) - project identity, layout, dependency policy.
- [200-go-style.md](200-go-style.md) - Go 1.21 compatibility and coding style.
- [300-testing.md](300-testing.md) - test organization and expectations.
- [400-docs.md](400-docs.md) - Go docs, README, plans, reports.
- [500-validation-and-workflow.md](500-validation-and-workflow.md) - validation gates and Make targets.
- [550-git-conventions.md](550-git-conventions.md) - commits, branches, no attribution trailers.

