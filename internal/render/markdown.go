package render

import (
	"fmt"
	"io"

	"github.com/arloliu/gocensus"
)

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
