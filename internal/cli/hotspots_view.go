package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/arloliu/gocensus/internal/color"
	"github.com/arloliu/gocensus/internal/hotspot"
)

func renderHotspots(w io.Writer, report hotspot.Report, format string, output string, style color.Style) error {
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
		if err := renderHotspotsTable(writer, report, style); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderHotspotsMarkdown(writer, report); err != nil {
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

func renderHotspotsTable(w io.Writer, report hotspot.Report, style color.Style) error {
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Title("Hotspots"), report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Scope"), report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%s: %s\n\n", style.Label("Sorted by"), style.Metric(report.SortBy)); err != nil {
		return err
	}

	table := textTable{
		Indent:  "  ",
		Gap:     "  ",
		Columns: hotspotsTableColumns(style),
		Rows:    hotspotsRows(report.Files, style),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Notes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "\n%s\n", style.Section("Notes:")); err != nil {
		return err
	}
	for _, note := range report.Notes {
		if _, err := fmt.Fprintf(w, "  %s\n", style.Muted(note)); err != nil {
			return err
		}
	}
	return nil
}

func hotspotsTableColumns(style color.Style) []tableColumn {
	return []tableColumn{
		{Header: style.Header("File"), Align: tableAlignLeft},
		{Header: style.Header("Score"), Align: tableAlignRight},
		{Header: style.Header("Eff Lines"), Align: tableAlignRight},
		{Header: style.Header("Churn"), Align: tableAlignRight},
		{Header: style.Header("Added"), Align: tableAlignRight},
		{Header: style.Header("Removed"), Align: tableAlignRight},
		{Header: style.Header("Commits"), Align: tableAlignRight},
		{Header: style.Header("Package Test Ratio"), Align: tableAlignRight},
	}
}

func hotspotsRows(files []hotspot.File, style color.Style) [][]string {
	rows := make([][]string, 0, len(files))
	for _, file := range files {
		rows = append(rows, []string{
			style.Label(file.Path),
			style.Metric(formatWhoInt(file.Score)),
			style.Metric(formatWhoInt(file.EffectiveLines)),
			style.Metric(formatWhoInt(file.Churn)),
			style.Metric(formatWhoInt(file.Added)),
			style.Warn(formatWhoInt(file.Removed)),
			style.Metric(formatWhoInt(file.Commits)),
			style.Metric(formatHotspotRatio(file.PackageTestRatio)),
		})
	}
	return rows
}

func renderHotspotsMarkdown(w io.Writer, report hotspot.Report) error {
	if _, err := fmt.Fprintf(w, "# Hotspots: %s\n\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: `%s`\n\n", report.SortBy); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| File | Score | Eff Lines | Churn | Added | Removed | Commits | Package Test Ratio |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, row := range hotspotsRows(report.Files, color.Plain()) {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7]); err != nil {
			return err
		}
	}
	if len(report.Notes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\n## Notes"); err != nil {
		return err
	}
	for _, note := range report.Notes {
		if _, err := fmt.Fprintf(w, "- %s\n", note); err != nil {
			return err
		}
	}
	return nil
}

func formatHotspotRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
