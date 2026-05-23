package report

import (
	"slices"

	"github.com/arloliu/gocensus"
)

// Input contains the file metrics to aggregate into a Result.
type Input struct {
	Root        string
	ModulePath  string
	FileMetrics []gocensus.FileMetric
}

// Build aggregates file metrics into repository and package metrics.
func Build(input Input) gocensus.Result {
	result := gocensus.Result{
		Root:        input.Root,
		ModulePath:  input.ModulePath,
		FileMetrics: slices.Clone(input.FileMetrics),
	}

	byPackage := map[string]*gocensus.PackageMetric{}
	for _, file := range input.FileMetrics {
		addFile(&result.Files, &result.Lines, &result.Tests, file)
		pkg := byPackage[file.Package]
		if pkg == nil {
			pkg = &gocensus.PackageMetric{
				ImportPath: file.Package,
				Dir:        file.Package,
			}
			byPackage[file.Package] = pkg
		}
		addFile(&pkg.Files, &pkg.Lines, &pkg.Tests, file)
	}
	result.Ratios = ratios(result.Lines)

	for _, pkg := range byPackage {
		pkg.Ratios = ratios(pkg.Lines)
		result.Packages = append(result.Packages, *pkg)
	}
	slices.SortFunc(result.Packages, func(a, b gocensus.PackageMetric) int {
		if a.ImportPath < b.ImportPath {
			return -1
		}
		if a.ImportPath > b.ImportPath {
			return 1
		}
		return 0
	})
	return result
}

func addFile(files *gocensus.FileCounts, lines *gocensus.LineCounts, tests *gocensus.TestCounts, file gocensus.FileMetric) {
	files.Total++
	tests.Tests += file.Tests
	tests.StaticSubtests += file.StaticSubtests
	tests.DynamicSubtestSites += file.DynamicSubtestSites
	tests.Benchmarks += file.Benchmarks
	tests.StaticSubbenchmarks += file.StaticSubbenchmarks
	tests.DynamicSubbenchmarkSites += file.DynamicSubbenchmarkSites
	tests.Examples += file.Examples

	metric := gocensus.Metric{Raw: file.RawLines, Effective: file.CodeLines}
	switch file.Kind {
	case "production":
		files.Production++
		lines.Production.Raw += metric.Raw
		lines.Production.Effective += metric.Effective
	case "test":
		files.Tests++
		lines.Tests.Raw += metric.Raw
		lines.Tests.Effective += metric.Effective
	case "generated":
		files.Generated++
		lines.Generated.Raw += metric.Raw
		lines.Generated.Effective += metric.Effective
	case "mock":
		files.Mocks++
		lines.Mocks.Raw += metric.Raw
		lines.Mocks.Effective += metric.Effective
	}
}

func ratios(lines gocensus.LineCounts) gocensus.Ratios {
	totalRaw := lines.Production.Raw + lines.Tests.Raw + lines.Generated.Raw + lines.Mocks.Raw
	totalEffective := lines.Production.Effective + lines.Tests.Effective
	return gocensus.Ratios{
		TestToProductionRaw:       divide(lines.Tests.Raw, lines.Production.Raw),
		TestToProductionEffective: divide(lines.Tests.Effective, lines.Production.Effective),
		TestShareEffective:        divide(lines.Tests.Effective, totalEffective),
		GeneratedShareRaw:         divide(lines.Generated.Raw, totalRaw),
		MockShareRaw:              divide(lines.Mocks.Raw, totalRaw),
	}
}

func divide(numerator int, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
