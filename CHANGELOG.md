# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Added

- Add `gocensus check` for CI-friendly repository policy checks, starting with `--min-test-ratio`.

## v0.4.0 - 2026-05-30

### Changed

- Exclude Go `testdata/` fixture directories from scan/report analysis by default, with `--include-testdata` to include them explicitly.

## v0.3.0 - 2026-05-25

### Added

- Add terminal color support for human-readable table output, with `--color` and `--no-color` controls.
- Add `gocensus diff` for comparing scan metrics between two Git refs without mutating the working tree.
- Add `gocensus hotspots` for ranking production Go file hotspots by effective lines plus Git churn.
- Add table, JSON, Markdown, and `--output` support for `diff` and `hotspots`.
- Add internal Git archive, diff, and hotspot analysis packages.

### Changed

- Soften the RGB terminal color palette for better readability.
- Make `gocensus who` follow scan-style generated/mock defaults: generated and mock paths are excluded by default and can be included with `--include-generated` or `--include-mocks`.
- Update README documentation for `diff`, `hotspots`, color output, and contributor-ranking scope behavior.

## v0.2.0 - 2026-05-24

### Added

- Add `gocensus who` for ranking Git contributors by commits, message-classified feature/fix/refactor counts, line changes, churn, files touched, and active days.
- Add `--go-only`, `--by`, `--since`, `--until`, `-n`, `--format`, and `--output` support for contributor reports.
- Add contributor report rendering in table, JSON, and Markdown formats.
- Add internal Git contribution analysis, parsing, filtering, and ranking support.
- Add a reusable CLI text-table renderer for aligned tabular output.
- Add a future implementation plan for `diff` and `hotspots` commands.

### Changed

- Clarify scan/report scope language so production totals explicitly describe generated and mock file handling.
- Exclude generated and mock buckets from default included file totals and production/test ratio totals unless explicitly included.
- Show `Scope:` in table and Markdown scan/report output.
- Rename rendered production labels to `Production Scope` and generated/mock labels to `Excluded Generated` and `Excluded Mocks`.
- Remove generated-share and mock-share ratios from the default rendered scan/report output.
- Update README examples, metric descriptions, and command documentation for the new scope language and contributor rankings.

## v0.1.0 - 2026-05-24

### Added

- Initial Go library and CLI for scanning Go repositories.
- Add repository discovery, Go file classification, line counting, test inventory, aggregation, and report rendering.
- Add `scan`, `report`, `packages`, `files`, `tests`, and `version` commands.
