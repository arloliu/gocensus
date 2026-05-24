# Changelog

All notable changes to this project are documented in this file.

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
