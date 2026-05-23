package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/render"
)

func renderResult(w io.Writer, result gocensus.Result, format string, output string) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	if err := render.Result(writer, result, format); err != nil {
		return fmt.Errorf("render result: %w", err)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderPackages(w io.Writer, result gocensus.Result, sortBy string) error {
	if _, err := fmt.Fprintln(w, "Packages"); err != nil {
		return err
	}
	packages := sortedPackages(result.Packages, sortBy)
	for _, pkg := range packages {
		if _, err := fmt.Fprintf(w, "  %s  prod=%d  test=%d  ratio=%.2f:1\n",
			pkg.ImportPath,
			pkg.Lines.Production.Effective,
			pkg.Lines.Tests.Effective,
			pkg.Ratios.TestToProductionEffective,
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

func renderFiles(w io.Writer, result gocensus.Result, top int) error {
	if _, err := fmt.Fprintln(w, "Files"); err != nil {
		return err
	}
	for i, file := range result.FileMetrics {
		if top > 0 && i >= top {
			break
		}
		if _, err := fmt.Fprintf(w, "  %s  %s  raw=%d  effective=%d\n",
			file.Path, file.Kind, file.RawLines, file.CodeLines); err != nil {
			return err
		}
	}
	return nil
}

func renderTests(w io.Writer, result gocensus.Result) error {
	knownTestCases := result.Tests.Tests + result.Tests.StaticSubtests
	knownBenchmarkCases := result.Tests.Benchmarks + result.Tests.StaticSubbenchmarks

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
