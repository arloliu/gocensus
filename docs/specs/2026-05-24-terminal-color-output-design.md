# Terminal Color Output Design

## Goal

Add modern, colorful terminal output across every human-readable gocensus view while preserving plain, stable output for machine formats, redirected artifacts, and users who opt out.

## Scope

Color applies to terminal-facing text for:

- `scan` table output.
- `report -f table` when writing to stdout.
- `packages`.
- `files`.
- `tests`.
- `who -f table`.

Color does not apply to:

- JSON output.
- Markdown output.
- Any command using `--output PATH`.
- Plain mode selected by flags or environment.

## CLI Controls

Add global color controls:

- `--color auto|always|never`, default `auto`.
- `--no-color` as a convenience alias for `--color never`.

Mode precedence:

1. `--no-color` disables color.
2. `--color never` disables color.
3. `--color always` forces color.
4. In `auto`, `NO_COLOR` disables color.
5. In `auto`, `CLICOLOR=0` disables color.
6. In `auto`, `CLICOLOR_FORCE` forces color when set to a non-empty value other than `0`.
7. Otherwise, auto mode colors stdout when terminal capability signals are present.

## Capability Detection

Use standard library logic only. Detect RGB-capable terminals from common environment signals:

- `COLORTERM=truecolor`.
- `COLORTERM=24bit`.
- Known modern terminal names in `TERM_PROGRAM`, such as iTerm.app, WezTerm, kitty, Hyper, Apple_Terminal, VSCode, Ghostty, and Windows Terminal.
- `TERM` values containing `truecolor` or `24bit`.

When RGB is unavailable but color is enabled, use basic ANSI SGR colors. When no color is enabled, emit no escape sequences.

## Styling

Define a small internal style layer with three levels:

- `plain`: no escape sequences.
- `ansi`: basic SGR codes such as bold, dim, cyan, green, yellow, red, and magenta.
- `rgb`: 24-bit SGR colors with a restrained modern palette.

Suggested styling:

- Titles: bold cyan.
- Section headings: bold blue/cyan.
- Table headers: bold muted cyan.
- Labels: muted cyan.
- Important totals and positive metrics: green.
- Warnings or excluded buckets: yellow.
- Negative contributor net values: red.
- Notes and secondary labels: dim.

The implementation should avoid changing words, spacing, or line order in plain mode.

## Renderer Boundaries

Keep analysis logic unchanged. Thread a color/style value only through CLI rendering:

- Add render options to `internal/render` for `scan` and `report` table output.
- Keep JSON and Markdown renderers plain.
- Add style support to `internal/cli` renderers for `packages`, `files`, `tests`, and `who`.
- Preserve file output behavior by resolving style to plain whenever `--output PATH` is set.

## Width And Alignment

ANSI escape sequences must not affect table width calculations. Update the CLI text table display-width helper to skip SGR control sequences before measuring visible characters. This keeps `who -f table` aligned when cells or headers are colored.

## Tests

Add tests before implementation for:

- Color mode resolution, including `--no-color`, `NO_COLOR`, `CLICOLOR=0`, `CLICOLOR_FORCE`, RGB signals, and ANSI fallback.
- Plain mode output contains no SGR sequences.
- Forced RGB table output contains 24-bit SGR sequences.
- Output written through `--output PATH` remains plain.
- SGR-decorated cells do not break table alignment.
- Each human-readable command family can render with color enabled without changing machine formats.

## Documentation

Update README command documentation to mention:

- Default auto color behavior.
- `--color auto|always|never`.
- `--no-color`.
- JSON, Markdown, and output files remain plain.
