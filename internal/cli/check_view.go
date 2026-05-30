package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/arloliu/gocensus/internal/check"
	"github.com/arloliu/gocensus/internal/color"
)

func renderCheck(w io.Writer, report check.Report, format string, output string, style color.Style) error {
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
		if err := renderCheckTable(writer, report, style); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderCheckMarkdown(writer, report); err != nil {
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

func renderCheckTable(w io.Writer, report check.Report, style color.Style) error {
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Title("Check"), checkDisplayName(report)); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Scope"), report.Scope); err != nil {
			return err
		}
	}
	status := "PASS"
	statusText := style.Metric(status)
	if !report.Passed {
		status = "FAIL"
		statusText = style.Bad(status)
	}
	if _, err := fmt.Fprintf(w, "%s: %s\n\n", style.Label("Status"), statusText); err != nil {
		return err
	}
	table := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: style.Header("Check"), Align: tableAlignLeft},
			{Header: style.Header("Status"), Align: tableAlignLeft},
			{Header: style.Header("Actual"), Align: tableAlignRight},
			{Header: style.Header("Required"), Align: tableAlignRight},
			{Header: style.Header("Message"), Align: tableAlignLeft},
		},
		Rows: checkRows(report.Checks, style),
	}
	return renderTextTable(w, table)
}

func renderCheckMarkdown(w io.Writer, report check.Report) error {
	if _, err := fmt.Fprintf(w, "# Check: %s\n\n", checkDisplayName(report)); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	status := "PASS"
	if !report.Passed {
		status = "FAIL"
	}
	if _, err := fmt.Fprintf(w, "Status: **%s**\n\n", status); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Check | Status | Actual | Required | Message |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | --- | ---: | ---: | --- |"); err != nil {
		return err
	}
	for _, row := range checkRows(report.Checks, color.Plain()) {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3], row[4]); err != nil {
			return err
		}
	}
	return nil
}

func checkRows(checks []check.Check, style color.Style) [][]string {
	rows := make([][]string, 0, len(checks))
	for _, item := range checks {
		status := "PASS"
		statusText := style.Metric(status)
		if !item.Passed {
			status = "FAIL"
			statusText = style.Bad(status)
		}
		rows = append(rows, []string{
			style.Label(item.Label),
			statusText,
			style.Metric(formatCheckRatio(item.Actual)),
			style.Metric(formatCheckRatio(item.Threshold)),
			style.Muted(item.Message),
		})
	}
	return rows
}

func checkDisplayName(report check.Report) string {
	if report.ModulePath != "" {
		return report.ModulePath
	}
	return report.Root
}

func formatCheckRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
