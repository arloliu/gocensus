package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/arloliu/gocensus/internal/color"
	censusdiff "github.com/arloliu/gocensus/internal/diff"
)

func renderDiff(w io.Writer, report censusdiff.Report, format string, output string, style color.Style) error {
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
		if err := renderDiffTable(writer, report, style); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderDiffMarkdown(writer, report); err != nil {
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

func renderDiffTable(w io.Writer, report censusdiff.Report, style color.Style) error {
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Title("Diff"), report.Root); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Base"), style.Metric(report.Base)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Head"), style.Metric(report.Head)); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Scope"), report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "\n%s\n", style.Section("Summary")); err != nil {
		return err
	}
	table := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: style.Header("Metric"), Align: tableAlignLeft},
			{Header: style.Header("Base"), Align: tableAlignRight},
			{Header: style.Header("Head"), Align: tableAlignRight},
			{Header: style.Header("Delta"), Align: tableAlignRight},
		},
		Rows: diffSummaryRows(report.Summary, style),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Packages) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "\n%s\n", style.Section("Packages")); err != nil {
		return err
	}
	packageTable := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: style.Header("Package"), Align: tableAlignLeft},
			{Header: style.Header("Prod Eff"), Align: tableAlignRight},
			{Header: style.Header("Test Eff"), Align: tableAlignRight},
			{Header: style.Header("Ratio"), Align: tableAlignRight},
			{Header: style.Header("Prod Delta"), Align: tableAlignRight},
			{Header: style.Header("Test Delta"), Align: tableAlignRight},
		},
		Rows: diffPackageRows(report.Packages, style),
	}
	return renderTextTable(w, packageTable)
}

func renderDiffMarkdown(w io.Writer, report censusdiff.Report) error {
	if _, err := fmt.Fprintf(w, "# Diff: %s\n\n", report.Root); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Base: `%s`\n\nHead: `%s`\n\n", report.Base, report.Head); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "| Metric | Base | Head | Delta |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, row := range diffSummaryRows(report.Summary, color.Plain()) {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3]); err != nil {
			return err
		}
	}
	return nil
}

func diffSummaryRows(summary censusdiff.Summary, style color.Style) [][]string {
	return [][]string{
		{style.Label("Go Files"), style.Metric(formatWhoInt(summary.TotalFiles.Base)), style.Metric(formatWhoInt(summary.TotalFiles.Head)), styleDiffInt(summary.TotalFiles.Delta, style)},
		{style.Label("Production Files"), style.Metric(formatWhoInt(summary.ProductionFiles.Base)), style.Metric(formatWhoInt(summary.ProductionFiles.Head)), styleDiffInt(summary.ProductionFiles.Delta, style)},
		{style.Label("Test Files"), style.Metric(formatWhoInt(summary.TestFiles.Base)), style.Metric(formatWhoInt(summary.TestFiles.Head)), styleDiffInt(summary.TestFiles.Delta, style)},
		{style.Label("Production Effective"), style.Metric(formatWhoInt(summary.ProductionEffective.Base)), style.Metric(formatWhoInt(summary.ProductionEffective.Head)), styleDiffInt(summary.ProductionEffective.Delta, style)},
		{style.Label("Test Effective"), style.Metric(formatWhoInt(summary.TestEffective.Base)), style.Metric(formatWhoInt(summary.TestEffective.Head)), styleDiffInt(summary.TestEffective.Delta, style)},
		{style.Label("Test / Production Scope"), style.Metric(formatDiffRatio(summary.TestToProduction.Base)), style.Metric(formatDiffRatio(summary.TestToProduction.Head)), styleDiffFloat(summary.TestToProduction.Delta, style)},
		{style.Label("Test Share"), style.Metric(formatDiffPercent(summary.TestShare.Base)), style.Metric(formatDiffPercent(summary.TestShare.Head)), styleDiffPercent(summary.TestShare.Delta, style)},
	}
}

func diffPackageRows(packages []censusdiff.PackageDelta, style color.Style) [][]string {
	rows := make([][]string, 0, len(packages))
	for _, pkg := range packages {
		rows = append(rows, []string{
			style.Label(pkg.Package),
			style.Metric(formatWhoInt(pkg.ProductionEffective.Head)),
			style.Metric(formatWhoInt(pkg.TestEffective.Head)),
			style.Metric(formatDiffRatio(pkg.TestToProduction.Head)),
			styleDiffInt(pkg.ProductionEffective.Delta, style),
			styleDiffInt(pkg.TestEffective.Delta, style),
		})
	}
	return rows
}

func styleDiffInt(value int, style color.Style) string {
	text := signedInt(value)
	if value < 0 {
		return style.Bad(text)
	}
	return style.Metric(text)
}

func signedInt(value int) string {
	if value > 0 {
		return "+" + formatWhoInt(value)
	}
	return formatWhoInt(value)
}

func formatDiffRatio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}

func formatDiffPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func signedFloat(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.2f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func styleDiffFloat(value float64, style color.Style) string {
	text := signedFloat(value)
	if value < 0 {
		return style.Bad(text)
	}
	return style.Metric(text)
}

func signedPercent(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.1f%%", value*100)
	}
	return fmt.Sprintf("%.1f%%", value*100)
}

func styleDiffPercent(value float64, style color.Style) string {
	text := signedPercent(value)
	if value < 0 {
		return style.Bad(text)
	}
	return style.Metric(text)
}
