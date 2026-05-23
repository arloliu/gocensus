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
	_, err := fmt.Fprintf(w, `Go Census: %s

Files
  Production %d
  Tests      %d
  Generated  %d
  Mocks      %d
  Total      %d

Lines
  Production %d raw  %d effective
  Tests      %d raw  %d effective
  Generated  %d raw  %d effective
  Mocks      %d raw  %d effective

Ratios
  Test / Production %s effective
  Test Share        %s
  Generated Share   %s
  Mock Share        %s

Tests
  Test funcs  %d
  Benchmarks  %d
  Examples    %d
`,
		displayName(result),
		result.Files.Production,
		result.Files.Tests,
		result.Files.Generated,
		result.Files.Mocks,
		result.Files.Total,
		result.Lines.Production.Raw,
		result.Lines.Production.Effective,
		result.Lines.Tests.Raw,
		result.Lines.Tests.Effective,
		result.Lines.Generated.Raw,
		result.Lines.Generated.Effective,
		result.Lines.Mocks.Raw,
		result.Lines.Mocks.Effective,
		ratio(result.Ratios.TestToProductionEffective),
		pct(result.Ratios.TestShareEffective),
		pct(result.Ratios.GeneratedShareRaw),
		pct(result.Ratios.MockShareRaw),
		result.Tests.Tests,
		result.Tests.Benchmarks,
		result.Tests.Examples,
	)
	return err
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
- Benchmarks: %d
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
		result.Tests.Benchmarks,
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

func pct(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func ratio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
