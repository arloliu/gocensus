package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/arloliu/gocensus/internal/color"
	"github.com/arloliu/gocensus/internal/contrib"
)

func renderWho(w io.Writer, report contrib.Report, sortBy string, format string, output string, style color.Style) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	if output != "" || format != "table" {
		style = color.Plain()
	}
	switch format {
	case "table":
		if err := renderWhoTable(writer, report, sortBy, style); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderWhoMarkdown(writer, report, sortBy); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderWhoTable(w io.Writer, report contrib.Report, sortBy string, style color.Style) error {
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Title("Who"), report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Scope"), report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%s: %s\n\n", style.Label("Sorted by"), style.Metric(sortBy)); err != nil {
		return err
	}

	table := textTable{
		Indent:  "  ",
		Gap:     "  ",
		Columns: whoTableColumns(style),
		Rows:    whoRows(report.Contributors, style),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Notes) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, style.Section("Notes:")); err != nil {
			return err
		}
		for _, note := range report.Notes {
			if _, err := fmt.Fprintf(w, "  %s\n", style.Muted(displayWhoNote(note))); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderWhoMarkdown(w io.Writer, report contrib.Report, sortBy string) error {
	if _, err := fmt.Fprintf(w, "# Who: %s\n\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: `%s`\n\n", sortBy); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Author | Commits | Feat | Fix | Refactor | Added | Removed | Net | Churn | Files | Active Days |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, contributor := range report.Contributors {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			contributor.Name,
			formatWhoInt(contributor.Commits),
			formatWhoInt(contributor.Features),
			formatWhoInt(contributor.Fixes),
			formatWhoInt(contributor.Refactors),
			formatWhoInt(contributor.Added),
			formatWhoInt(contributor.Removed),
			formatWhoInt(contributor.Net),
			formatWhoInt(contributor.Churn),
			formatWhoInt(contributor.Files),
			formatWhoInt(contributor.ActiveDays),
		); err != nil {
			return err
		}
	}
	if len(report.Notes) > 0 {
		if _, err := fmt.Fprintln(w, "\n## Notes"); err != nil {
			return err
		}
		for _, note := range report.Notes {
			if _, err := fmt.Fprintf(w, "- %s\n", displayWhoNote(note)); err != nil {
				return err
			}
		}
	}
	return nil
}

func whoTableColumns(style color.Style) []tableColumn {
	return []tableColumn{
		{Header: style.Header("Author"), Align: tableAlignLeft},
		{Header: style.Header("Commits"), Align: tableAlignRight},
		{Header: style.Header("Feat"), Align: tableAlignRight},
		{Header: style.Header("Fix"), Align: tableAlignRight},
		{Header: style.Header("Refactor"), Align: tableAlignRight},
		{Header: style.Header("Added"), Align: tableAlignRight},
		{Header: style.Header("Removed"), Align: tableAlignRight},
		{Header: style.Header("Net"), Align: tableAlignRight},
		{Header: style.Header("Churn"), Align: tableAlignRight},
		{Header: style.Header("Files"), Align: tableAlignRight},
		{Header: style.Header("Active Days"), Align: tableAlignRight},
	}
}

func whoRows(contributors []contrib.Contributor, style color.Style) [][]string {
	rows := make([][]string, 0, len(contributors))
	for _, contributor := range contributors {
		rows = append(rows, []string{
			style.Label(contributor.Name),
			style.Metric(formatWhoInt(contributor.Commits)),
			style.Metric(formatWhoInt(contributor.Features)),
			style.Metric(formatWhoInt(contributor.Fixes)),
			style.Metric(formatWhoInt(contributor.Refactors)),
			style.Metric(formatWhoInt(contributor.Added)),
			style.Warn(formatWhoInt(contributor.Removed)),
			styleNet(contributor.Net, style),
			style.Metric(formatWhoInt(contributor.Churn)),
			style.Metric(formatWhoInt(contributor.Files)),
			style.Metric(formatWhoInt(contributor.ActiveDays)),
		})
	}
	return rows
}

func styleNet(value int, style color.Style) string {
	text := formatWhoInt(value)
	if value < 0 {
		return style.Bad(text)
	}
	return style.Metric(text)
}

func formatWhoInt(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	text := strconv.Itoa(value)
	if len(text) <= 3 {
		return sign + text
	}

	firstGroup := len(text) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}

	var out strings.Builder
	out.WriteString(text[:firstGroup])
	for i := firstGroup; i < len(text); i += 3 {
		out.WriteString("," + text[i:i+3])
	}
	return sign + out.String()
}

func displayWhoNote(note string) string {
	switch note {
	case "Feature, fix, and refactor counts are commit-message heuristics.":
		return "Feature/Fix/Refactor: commit-message heuristics."
	case "Line, file, commit, and active-day counts come from git log --numstat.":
		return "Line metrics: git log --numstat."
	default:
		return note
	}
}
