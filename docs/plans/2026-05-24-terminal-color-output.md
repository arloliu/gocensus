# Terminal Color Output Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add RGB/ANSI/plain color support for every human-readable terminal output while keeping machine formats and file output plain.

**Architecture:** Introduce a small `internal/color` package that resolves color mode from CLI flags and environment and exposes style helpers. Thread the resolved style through `internal/cli` renderers and `internal/render` table rendering without changing analysis logic.

**Tech Stack:** Go 1.21, standard library only, existing Kong CLI parser, existing unit tests.

---

## File Structure

- Create `internal/color/color.go`: color mode, terminal capability detection, SGR style helpers, and SGR stripping/width support.
- Create `internal/color/color_test.go`: mode-resolution and style helper tests.
- Modify `internal/cli/commands.go`: add global `--color` and `--no-color`, pass resolved style into renderers.
- Modify `internal/cli/cli.go`: pass `os.Environ()` into runtime so tests can control color decisions.
- Modify `internal/cli/views.go`: color `packages`, `files`, and `tests` human output and keep output files plain.
- Modify `internal/cli/who_view.go`: color table title, scope, sort label, notes, and key numeric cells.
- Modify `internal/cli/table.go`: measure display width after stripping SGR control sequences.
- Modify `internal/cli/table_test.go`: add a colored-cell alignment regression test.
- Modify `internal/render/render.go`: add `Options` and `ResultWithOptions` while preserving existing `Result`.
- Modify `internal/render/table.go`: color scan/report table headings, labels, totals, ratios, warnings, and notes.
- Modify `internal/render/render_test.go`: keep existing plain snapshot stable and add forced RGB table test.
- Modify `internal/cli/cli_test.go` and `internal/cli/views_test.go`: add CLI-level flag/file-output coverage.
- Modify `internal/cli/help_test.go`: update help assertions for color flags.
- Modify `README.md`: document color controls.

## Tasks

1. Add `internal/color` with tests for mode resolution and SGR stripping.
2. Add render options and color the main scan/report table.
3. Add global CLI color flags and thread resolved style through all human renderers.
4. Make text-table width calculation ignore SGR sequences.
5. Update help tests and README docs.
6. Run `gofmt`, focused package tests, and `make check`.
