# AGENTS.md

## Project

`gocensus` is a Go library and CLI for analyzing Go repositories: files, lines, tests, ratios, and reports.

Target Go compatibility is **Go 1.21+**. Do not introduce language features, standard-library APIs, or dependencies that raise that floor unless explicitly requested.

## Rule Files

Repository rules live in [.agents/rules/](.agents/rules/). Start with [.agents/rules/AGENTS.md](.agents/rules/AGENTS.md), then read the matching rule files for the task.

For ordinary Go changes, read:

- [.agents/rules/000-agent-contract.md](.agents/rules/000-agent-contract.md)
- [.agents/rules/100-project-map.md](.agents/rules/100-project-map.md)
- [.agents/rules/200-go-style.md](.agents/rules/200-go-style.md)
- [.agents/rules/300-testing.md](.agents/rules/300-testing.md)
- [.agents/rules/500-validation-and-workflow.md](.agents/rules/500-validation-and-workflow.md)

## Workflow

- Keep changes small and focused.
- Prefer standard-library solutions.
- Keep analysis logic separate from CLI rendering.
- Update tests for behavior changes.
- Update README or Go docs when exported API or CLI behavior changes.

## Validation

Run the relevant narrow check while iterating. Before calling implementation work complete, run:

```bash
make check
```

Before a git commit, always run the linter and fix linting issues.

## Git

- Use Conventional Commits for commit messages.
- Never add `Co-Authored-By` or any other attribution trailers.

