package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/color"
	"github.com/arloliu/gocensus/internal/render"
)

func renderResult(w io.Writer, result gocensus.Result, format string, output string, style color.Style) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	if output != "" || format != "table" {
		style = color.Plain()
	}
	if err := render.ResultWithOptions(writer, result, format, render.Options{Style: style}); err != nil {
		return fmt.Errorf("render result: %w", err)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderPackages(w io.Writer, result gocensus.Result, sortBy string, style color.Style) error {
	if _, err := fmt.Fprintln(w, style.Section("Packages")); err != nil {
		return err
	}
	packages := sortedPackages(result.Packages, sortBy)
	for _, pkg := range packages {
		if _, err := fmt.Fprintf(w, "  %s  %s=%s  %s=%s  %s=%s\n",
			style.Label(pkg.ImportPath),
			style.Muted("prod"),
			style.Metric(fmt.Sprintf("%d", pkg.Lines.Production.Effective)),
			style.Muted("test"),
			style.Metric(fmt.Sprintf("%d", pkg.Lines.Tests.Effective)),
			style.Muted("ratio"),
			style.Metric(fmt.Sprintf("%.2f:1", pkg.Ratios.TestToProductionEffective)),
		); err != nil {
			return err
		}
	}
	return nil
}

func sortedPackages(packages []gocensus.PackageMetric, sortBy string) []gocensus.PackageMetric {
	sorted := slices.Clone(packages)
	slices.SortFunc(sorted, func(a gocensus.PackageMetric, b gocensus.PackageMetric) int {
		switch sortBy {
		case "test-ratio":
			if a.Ratios.TestToProductionEffective > b.Ratios.TestToProductionEffective {
				return -1
			}
			if a.Ratios.TestToProductionEffective < b.Ratios.TestToProductionEffective {
				return 1
			}
		case "prod-lines":
			if a.Lines.Production.Effective > b.Lines.Production.Effective {
				return -1
			}
			if a.Lines.Production.Effective < b.Lines.Production.Effective {
				return 1
			}
		case "test-lines":
			if a.Lines.Tests.Effective > b.Lines.Tests.Effective {
				return -1
			}
			if a.Lines.Tests.Effective < b.Lines.Tests.Effective {
				return 1
			}
		}
		if a.ImportPath < b.ImportPath {
			return -1
		}
		if a.ImportPath > b.ImportPath {
			return 1
		}
		return 0
	})
	return sorted
}

func renderFiles(w io.Writer, result gocensus.Result, top int, style color.Style) error {
	if _, err := fmt.Fprintln(w, style.Section("Files")); err != nil {
		return err
	}
	for i, file := range result.FileMetrics {
		if top > 0 && i >= top {
			break
		}
		kind := style.Label(file.Kind)
		if file.Generated || file.Kind == "generated" || file.Kind == "mock" {
			kind = style.Warn(file.Kind)
		}
		if _, err := fmt.Fprintf(w, "  %s  %s  %s=%s  %s=%s\n",
			style.Label(file.Path),
			kind,
			style.Muted("raw"),
			style.Metric(fmt.Sprintf("%d", file.RawLines)),
			style.Muted("effective"),
			style.Metric(fmt.Sprintf("%d", file.CodeLines))); err != nil {
			return err
		}
	}
	return nil
}

func renderTests(w io.Writer, result gocensus.Result, style color.Style) error {
	knownTestCases := result.Tests.Tests + result.Tests.StaticSubtests
	knownBenchmarkCases := result.Tests.Benchmarks + result.Tests.StaticSubbenchmarks

	if style.Level() != color.LevelPlain {
		return renderTestsColored(w, result, style, knownTestCases, knownBenchmarkCases)
	}
	_, err := fmt.Fprintf(w, `Tests
  Known Test Cases         %d
  Top-level Tests          %d
  Static Subtests          %d
  Dynamic Subtest Sites    %d

Benchmarks
  Known Benchmark Cases    %d
  Top-level Benchmarks     %d
  Static Subbenchmarks     %d
  Dynamic Benchmark Sites  %d

Examples
  Examples                 %d
`,
		knownTestCases,
		result.Tests.Tests,
		result.Tests.StaticSubtests,
		result.Tests.DynamicSubtestSites,
		knownBenchmarkCases,
		result.Tests.Benchmarks,
		result.Tests.StaticSubbenchmarks,
		result.Tests.DynamicSubbenchmarkSites,
		result.Tests.Examples,
	)
	return err
}

func renderTestsColored(w io.Writer, result gocensus.Result, style color.Style, knownTestCases int, knownBenchmarkCases int) error {
	_, err := fmt.Fprintf(w, `%s
  %s         %s
  %s          %s
  %s          %s
  %s    %s

%s
  %s    %s
  %s     %s
  %s     %s
  %s  %s

%s
  %s                 %s
`,
		style.Section("Tests"),
		style.Label("Known Test Cases"),
		style.Metric(fmt.Sprintf("%d", knownTestCases)),
		style.Label("Top-level Tests"),
		style.Metric(fmt.Sprintf("%d", result.Tests.Tests)),
		style.Label("Static Subtests"),
		style.Metric(fmt.Sprintf("%d", result.Tests.StaticSubtests)),
		style.Label("Dynamic Subtest Sites"),
		style.Metric(fmt.Sprintf("%d", result.Tests.DynamicSubtestSites)),
		style.Section("Benchmarks"),
		style.Label("Known Benchmark Cases"),
		style.Metric(fmt.Sprintf("%d", knownBenchmarkCases)),
		style.Label("Top-level Benchmarks"),
		style.Metric(fmt.Sprintf("%d", result.Tests.Benchmarks)),
		style.Label("Static Subbenchmarks"),
		style.Metric(fmt.Sprintf("%d", result.Tests.StaticSubbenchmarks)),
		style.Label("Dynamic Benchmark Sites"),
		style.Metric(fmt.Sprintf("%d", result.Tests.DynamicSubbenchmarkSites)),
		style.Section("Examples"),
		style.Label("Examples"),
		style.Metric(fmt.Sprintf("%d", result.Tests.Examples)),
	)
	return err
}
