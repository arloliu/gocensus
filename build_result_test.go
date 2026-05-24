package gocensus

import (
	"reflect"
	"testing"
)

func TestBuildResultAggregatesRepoAndPackageMetrics(t *testing.T) {
	files := []FileMetric{
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

	result := buildResult("/repo", "example.com/app", files, Options{})

	if result.Files.Total != 2 {
		t.Fatalf("Files.Total = %d, want 2 included production/test files", result.Files.Total)
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
	if result.Ratios.GeneratedShareRaw != 0 || result.Ratios.MockShareRaw != 0 {
		t.Fatalf("excluded shares = generated %v mock %v, want 0/0 because excluded buckets are outside included totals",
			result.Ratios.GeneratedShareRaw, result.Ratios.MockShareRaw)
	}
	if len(result.Packages) != 1 {
		t.Fatalf("package count = %d, want 1", len(result.Packages))
	}
	if !reflect.DeepEqual(result.FileMetrics, files) {
		t.Fatalf("FileMetrics changed")
	}
}

func TestBuildResultIncludesGeneratedAndMocksWhenRequested(t *testing.T) {
	files := []FileMetric{
		{Path: "main.go", Package: "main", Kind: "production", RawLines: 10, CodeLines: 8},
		{Path: "main_test.go", Package: "main", Kind: "test", RawLines: 6, CodeLines: 4},
		{Path: "mock_client.go", Package: "main", Kind: "mock", RawLines: 3, CodeLines: 2},
		{Path: "service.pb.go", Package: "main", Kind: "generated", RawLines: 20, CodeLines: 18},
	}

	result := buildResult("/repo", "example.com/app", files, Options{
		IncludeGenerated: true,
		IncludeMocks:     true,
	})

	if result.Files.Total != 4 {
		t.Fatalf("Files.Total = %d, want all files included", result.Files.Total)
	}
	if result.Files.Production != 3 || result.Files.Tests != 1 || result.Files.Mocks != 0 || result.Files.Generated != 0 {
		t.Fatalf("file buckets = %#v", result.Files)
	}
	if result.Lines.Production.Effective != 28 {
		t.Fatalf("production effective = %d, want 28", result.Lines.Production.Effective)
	}
	if result.Ratios.TestToProductionEffective != float64(4)/28 {
		t.Fatalf("effective ratio = %v, want %v", result.Ratios.TestToProductionEffective, float64(4)/28)
	}
}
