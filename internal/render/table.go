package render

import (
	"fmt"
	"io"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/color"
)

func table(w io.Writer, result gocensus.Result, style color.Style) error {
	totalRaw := result.Lines.Production.Raw + result.Lines.Tests.Raw
	totalEffective := result.Lines.Production.Effective + result.Lines.Tests.Effective
	knownTestCases := result.Tests.Tests + result.Tests.StaticSubtests
	knownBenchmarkCases := result.Tests.Benchmarks + result.Tests.StaticSubbenchmarks

	if _, err := fmt.Fprintf(w, "%s: %s\n", style.Title("Go Census"), displayName(result)); err != nil {
		return err
	}
	if result.Scope != "" {
		if _, err := fmt.Fprintf(w, "%s: %s\n", style.Label("Scope"), result.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\n  %s: %s    %s: %s    %s: %s\n\n",
		style.Section("Overview"),
		style.Label("Go files"),
		style.Metric(formatInt(result.Files.Total)),
		style.Label("Packages"),
		style.Metric(formatInt(len(result.Packages))),
		style.Label("Known test cases"),
		style.Metric(formatInt(knownTestCases)),
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, style.Section("Code Mix")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s   %s   %s\n",
		style.Header(fmt.Sprintf("%-21s", "Kind")),
		style.Header(fmt.Sprintf("%5s", "Files")),
		style.Header(fmt.Sprintf("%9s", "Raw Lines")),
		style.Header(fmt.Sprintf("%15s", "Effective Lines")),
	); err != nil {
		return err
	}
	if err := codeMixRow(w, style, "Production Scope", result.Files.Production, result.Lines.Production.Raw, result.Lines.Production.Effective, codeMixPrimary); err != nil {
		return err
	}
	if err := codeMixRow(w, style, "Tests", result.Files.Tests, result.Lines.Tests.Raw, result.Lines.Tests.Effective, codeMixMetric); err != nil {
		return err
	}
	if err := codeMixRow(w, style, "Excluded Generated", result.Files.Generated, result.Lines.Generated.Raw, result.Lines.Generated.Effective, codeMixWarn); err != nil {
		return err
	}
	if err := codeMixRow(w, style, "Excluded Mocks", result.Files.Mocks, result.Lines.Mocks.Raw, result.Lines.Mocks.Effective, codeMixWarn); err != nil {
		return err
	}
	if err := codeMixRow(w, style, "Total", result.Files.Total, totalRaw, totalEffective, codeMixMetric); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.Section("Ratios")); err != nil {
		return err
	}
	if err := ratioRow(w, style, "Test / Production Scope", ratio(result.Ratios.TestToProductionEffective)); err != nil {
		return err
	}
	if err := ratioRow(w, style, "Test Share", pct(result.Ratios.TestShareEffective)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.Section("Test Inventory")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Known Test Cases")), style.Metric(fmt.Sprintf("%3s", formatInt(knownTestCases)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Top-level Tests")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.Tests)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Static Subtests")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.StaticSubtests)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Dynamic Subtest Sites")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.DynamicSubtestSites)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Known Benchmark Cases")), style.Metric(fmt.Sprintf("%3s", formatInt(knownBenchmarkCases)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Benchmarks")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.Benchmarks)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Static Subbenchmarks")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.StaticSubbenchmarks)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Dynamic Benchmark Sites")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.DynamicSubbenchmarkSites)))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s%s\n", style.Label(fmt.Sprintf("%-23s", "Examples")), style.Metric(fmt.Sprintf("%3s", formatInt(result.Tests.Examples)))); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, style.Section("Notes")); err != nil {
		return err
	}
	notes := []struct {
		field string
		text  string
	}{
		{field: "Raw Lines", text: "Physical lines, including blanks and comments."},
		{field: "Effective Lines", text: "Lines containing non-comment Go tokens."},
		{field: "Production Scope", text: "Non-test Go files counted as production; scope line shows generated/mock inclusion."},
		{field: "Tests", text: "*_test.go files."},
		{field: "Known Cases", text: "Top-level tests plus statically countable subtests."},
		{field: "Static Subtests", text: "t.Run/b.Run cases with statically countable case data."},
		{field: "Dynamic Sites", text: "t.Run/b.Run call sites with runtime-dependent case counts."},
		{field: "Excluded Generated", text: "Generated files not counted in production scope."},
		{field: "Excluded Mocks", text: "Mock/support files not counted in production scope."},
	}
	for _, note := range notes {
		if _, err := fmt.Fprintf(w, "  %s %s\n", style.Muted(fmt.Sprintf("%-20s", note.field)), style.Muted(note.text)); err != nil {
			return err
		}
	}
	return nil
}

type codeMixStyle int

const (
	codeMixPrimary codeMixStyle = iota
	codeMixMetric
	codeMixWarn
)

func codeMixRow(w io.Writer, style color.Style, kind string, files int, raw int, effective int, rowStyle codeMixStyle) error {
	kindText := fmt.Sprintf("%-20s", kind)
	filesText := fmt.Sprintf("%5s", formatInt(files))
	rawText := fmt.Sprintf("%9s", formatInt(raw))
	effectiveText := fmt.Sprintf("%15s", formatInt(effective))
	switch rowStyle {
	case codeMixPrimary:
		kindText = style.Label(kindText)
		filesText = style.Metric(filesText)
		rawText = style.Metric(rawText)
		effectiveText = style.Metric(effectiveText)
	case codeMixWarn:
		kindText = style.Warn(kindText)
		filesText = style.Warn(filesText)
		rawText = style.Warn(rawText)
		effectiveText = style.Warn(effectiveText)
	default:
		kindText = style.Header(kindText)
		filesText = style.Metric(filesText)
		rawText = style.Metric(rawText)
		effectiveText = style.Metric(effectiveText)
	}
	_, err := fmt.Fprintf(w, "  %s%s   %s   %s\n", kindText, filesText, rawText, effectiveText)
	return err
}

func ratioRow(w io.Writer, style color.Style, label string, value string) error {
	_, err := fmt.Fprintf(w, "  %s %s\n", style.Label(fmt.Sprintf("%-28s", label)), style.Metric(fmt.Sprintf("%7s", value)))
	return err
}
