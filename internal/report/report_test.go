package report_test

import (
	"reflect"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/report"
)

func TestBuildAggregatesRepoAndPackageMetrics(t *testing.T) {
	files := []gocensus.FileMetric{
		{Path: "main.go", Package: "main", Kind: "production", RawLines: 10, CodeLines: 8},
		{
			Path:                     "main_test.go",
			Package:                  "main",
			Kind:                     "test",
			RawLines:                 6,
			CodeLines:                4,
			Tests:                    2,
			StaticSubtests:           3,
			DynamicSubtestSites:      1,
			Benchmarks:               1,
			StaticSubbenchmarks:      2,
			DynamicSubbenchmarkSites: 1,
		},
		{Path: "mock_client.go", Package: "main", Kind: "mock", RawLines: 3, CodeLines: 2},
		{Path: "service.pb.go", Package: "main", Kind: "generated", RawLines: 20, CodeLines: 18},
	}

	result := report.Build(report.Input{
		Root:        "/repo",
		ModulePath:  "example.com/app",
		FileMetrics: files,
	})

	if result.Files.Total != 4 {
		t.Fatalf("Files.Total = %d, want 4", result.Files.Total)
	}
	if result.Files.Production != 1 || result.Files.Tests != 1 || result.Files.Mocks != 1 || result.Files.Generated != 1 {
		t.Fatalf("file buckets = %#v", result.Files)
	}
	if result.Lines.Production.Raw != 10 || result.Lines.Tests.Raw != 6 {
		t.Fatalf("line buckets = %#v", result.Lines)
	}
	if result.Tests.Tests != 2 {
		t.Fatalf("Tests.Tests = %d, want 2", result.Tests.Tests)
	}
	if result.Tests.StaticSubtests != 3 {
		t.Fatalf("Tests.StaticSubtests = %d, want 3", result.Tests.StaticSubtests)
	}
	if result.Tests.DynamicSubtestSites != 1 {
		t.Fatalf("Tests.DynamicSubtestSites = %d, want 1", result.Tests.DynamicSubtestSites)
	}
	if result.Tests.StaticSubbenchmarks != 2 {
		t.Fatalf("Tests.StaticSubbenchmarks = %d, want 2", result.Tests.StaticSubbenchmarks)
	}
	if result.Tests.DynamicSubbenchmarkSites != 1 {
		t.Fatalf("Tests.DynamicSubbenchmarkSites = %d, want 1", result.Tests.DynamicSubbenchmarkSites)
	}
	if result.Ratios.TestToProductionEffective != 0.5 {
		t.Fatalf("effective ratio = %v, want 0.5", result.Ratios.TestToProductionEffective)
	}
	if len(result.Packages) != 1 {
		t.Fatalf("package count = %d, want 1", len(result.Packages))
	}
	if !reflect.DeepEqual(result.FileMetrics, files) {
		t.Fatalf("FileMetrics changed")
	}
}
