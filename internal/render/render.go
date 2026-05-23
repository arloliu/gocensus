package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/arloliu/gocensus"
)

// Result renders a census result in the requested format.
func Result(w io.Writer, result gocensus.Result, format string) error {
	switch format {
	case "table":
		return table(w, result)
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "markdown":
		return markdown(w, result)
	default:
		return errors.New("unknown format")
	}
}

func table(w io.Writer, result gocensus.Result) error {
	totalRaw := result.Lines.Production.Raw + result.Lines.Tests.Raw + result.Lines.Generated.Raw + result.Lines.Mocks.Raw
	totalEffective := result.Lines.Production.Effective + result.Lines.Tests.Effective + result.Lines.Generated.Effective + result.Lines.Mocks.Effective
	knownTestCases := result.Tests.Tests + result.Tests.StaticSubtests
	knownBenchmarkCases := result.Tests.Benchmarks + result.Tests.StaticSubbenchmarks

	if _, err := fmt.Fprintf(w, "Go Census: %s\n\n", displayName(result)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Overview\n  Go files: %s    Packages: %s    Known test cases: %s\n\n",
		formatInt(result.Files.Total),
		formatInt(len(result.Packages)),
		formatInt(knownTestCases),
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "Code Mix"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "  Kind          Files   Raw Lines   Effective Lines"); err != nil {
		return err
	}
	if err := codeMixRow(w, "Production", result.Files.Production, result.Lines.Production.Raw, result.Lines.Production.Effective); err != nil {
		return err
	}
	if err := codeMixRow(w, "Tests", result.Files.Tests, result.Lines.Tests.Raw, result.Lines.Tests.Effective); err != nil {
		return err
	}
	if err := codeMixRow(w, "Generated", result.Files.Generated, result.Lines.Generated.Raw, result.Lines.Generated.Effective); err != nil {
		return err
	}
	if err := codeMixRow(w, "Mocks", result.Files.Mocks, result.Lines.Mocks.Raw, result.Lines.Mocks.Effective); err != nil {
		return err
	}
	if err := codeMixRow(w, "Total", result.Files.Total, totalRaw, totalEffective); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Ratios"); err != nil {
		return err
	}
	if err := ratioRow(w, "Test / Production", ratio(result.Ratios.TestToProductionEffective)); err != nil {
		return err
	}
	if err := ratioRow(w, "Test Share", pct(result.Ratios.TestShareEffective)); err != nil {
		return err
	}
	if err := ratioRow(w, "Generated Share", pct(result.Ratios.GeneratedShareRaw)); err != nil {
		return err
	}
	if err := ratioRow(w, "Mock Share", pct(result.Ratios.MockShareRaw)); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Test Inventory"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Known Test Cases", formatInt(knownTestCases)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Top-level Tests", formatInt(result.Tests.Tests)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Static Subtests", formatInt(result.Tests.StaticSubtests)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Dynamic Subtest Sites", formatInt(result.Tests.DynamicSubtestSites)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Known Benchmark Cases", formatInt(knownBenchmarkCases)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Benchmarks", formatInt(result.Tests.Benchmarks)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Static Subbenchmarks", formatInt(result.Tests.StaticSubbenchmarks)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Dynamic Benchmark Sites", formatInt(result.Tests.DynamicSubbenchmarkSites)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %-23s%3s\n", "Examples", formatInt(result.Tests.Examples)); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Notes"); err != nil {
		return err
	}
	notes := []struct {
		field string
		text  string
	}{
		{field: "Raw Lines", text: "Physical lines, including blanks and comments."},
		{field: "Effective Lines", text: "Lines containing non-comment Go tokens."},
		{field: "Production", text: "Non-test Go files, excluding generated files and mocks."},
		{field: "Tests", text: "*_test.go files."},
		{field: "Known Cases", text: "Top-level tests plus statically countable subtests."},
		{field: "Static Subtests", text: "t.Run/b.Run cases with statically countable case data."},
		{field: "Dynamic Sites", text: "t.Run/b.Run call sites with runtime-dependent case counts."},
		{field: "Generated", text: "Files with generated-code markers or generated suffixes."},
		{field: "Mocks", text: "Files classified as mock/support code."},
	}
	for _, note := range notes {
		if _, err := fmt.Fprintf(w, "  %-16s %s\n", note.field, note.text); err != nil {
			return err
		}
	}
	return nil
}

func markdown(w io.Writer, result gocensus.Result) error {
	if _, err := fmt.Fprintf(w, `# Go Census: %s

## Files

| Kind | Files | Raw Lines | Effective Lines |
| --- | ---: | ---: | ---: |
| Production | %d | %d | %d |
| Tests | %d | %d | %d |
| Generated | %d | %d | %d |
| Mocks | %d | %d | %d |

## Ratios

- Test / Production: %s effective
- Test Share: %s
- Generated Share: %s
- Mock Share: %s

## Tests

- Test funcs: %d
- Static subtests: %d
- Dynamic subtest sites: %d
- Benchmarks: %d
- Static sub-benchmarks: %d
- Dynamic sub-benchmark sites: %d
- Examples: %d

## Packages

| Package | Prod Lines | Test Lines | Test Ratio |
| --- | ---: | ---: | ---: |
`,
		displayName(result),
		result.Files.Production,
		result.Lines.Production.Raw,
		result.Lines.Production.Effective,
		result.Files.Tests,
		result.Lines.Tests.Raw,
		result.Lines.Tests.Effective,
		result.Files.Generated,
		result.Lines.Generated.Raw,
		result.Lines.Generated.Effective,
		result.Files.Mocks,
		result.Lines.Mocks.Raw,
		result.Lines.Mocks.Effective,
		ratio(result.Ratios.TestToProductionEffective),
		pct(result.Ratios.TestShareEffective),
		pct(result.Ratios.GeneratedShareRaw),
		pct(result.Ratios.MockShareRaw),
		result.Tests.Tests,
		result.Tests.StaticSubtests,
		result.Tests.DynamicSubtestSites,
		result.Tests.Benchmarks,
		result.Tests.StaticSubbenchmarks,
		result.Tests.DynamicSubbenchmarkSites,
		result.Tests.Examples,
	); err != nil {
		return err
	}

	for _, pkg := range result.Packages {
		if _, err := fmt.Fprintf(w, "| %s | %d | %d | %s |\n",
			pkg.ImportPath,
			pkg.Lines.Production.Effective,
			pkg.Lines.Tests.Effective,
			ratio(pkg.Ratios.TestToProductionEffective),
		); err != nil {
			return err
		}
	}
	return nil
}

func displayName(result gocensus.Result) string {
	if result.ModulePath != "" {
		return result.ModulePath
	}
	return result.Root
}

func codeMixRow(w io.Writer, kind string, files int, raw int, effective int) error {
	_, err := fmt.Fprintf(w, "  %-13s%5s   %9s   %15s\n",
		kind,
		formatInt(files),
		formatInt(raw),
		formatInt(effective),
	)
	return err
}

func ratioRow(w io.Writer, label string, value string) error {
	_, err := fmt.Fprintf(w, "  %-22s %7s\n", label, value)
	return err
}

func formatInt(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	text := fmt.Sprintf("%d", value)
	if len(text) <= 3 {
		return sign + text
	}

	firstGroup := len(text) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}

	out := text[:firstGroup]
	for i := firstGroup; i < len(text); i += 3 {
		out += "," + text[i:i+3]
	}
	return sign + out
}

func pct(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func ratio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
