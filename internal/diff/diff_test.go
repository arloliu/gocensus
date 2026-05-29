package diff_test

import (
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/diff"
)

func TestCompareComputesRepositoryDeltas(t *testing.T) {
	base := gocensus.Result{
		Files: gocensus.FileCounts{Total: 2, Production: 1, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Effective: 100},
			Tests:      gocensus.Metric{Effective: 50},
		},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
	}
	head := gocensus.Result{
		Files: gocensus.FileCounts{Total: 3, Production: 2, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Effective: 160},
			Tests:      gocensus.Metric{Effective: 80},
		},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
	}

	report := diff.Compare(diff.Options{
		Root:  "/repo",
		Base:  "v1",
		Head:  "v2",
		Scope: "production excludes generated and mock files; testdata directories excluded",
	}, base, head)

	if report.Root != "/repo" || report.Base != "v1" || report.Head != "v2" {
		t.Fatalf("identity = %#v", report)
	}
	if report.Summary.TotalFiles.Delta != 1 {
		t.Fatalf("total delta = %d, want 1", report.Summary.TotalFiles.Delta)
	}
	if report.Summary.ProductionEffective.Delta != 60 {
		t.Fatalf("production effective delta = %d, want 60", report.Summary.ProductionEffective.Delta)
	}
	if report.Summary.TestEffective.Delta != 30 {
		t.Fatalf("test effective delta = %d, want 30", report.Summary.TestEffective.Delta)
	}
}

func TestCompareSortsPackageDeltasByLargestProductionChange(t *testing.T) {
	base := gocensus.Result{
		Packages: []gocensus.PackageMetric{
			{
				ImportPath: "example.com/app/a",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 100},
					Tests:      gocensus.Metric{Effective: 20},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.2},
			},
			{
				ImportPath: "example.com/app/b",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 50},
					Tests:      gocensus.Metric{Effective: 10},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.2},
			},
		},
	}
	head := gocensus.Result{
		Packages: []gocensus.PackageMetric{
			{
				ImportPath: "example.com/app/a",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 75},
					Tests:      gocensus.Metric{Effective: 25},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.3333333333},
			},
			{
				ImportPath: "example.com/app/b",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 90},
					Tests:      gocensus.Metric{Effective: 10},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.1111111111},
			},
		},
	}

	report := diff.Compare(diff.Options{}, base, head)

	if len(report.Packages) != 2 {
		t.Fatalf("package count = %d, want 2", len(report.Packages))
	}
	if report.Packages[0].Package != "example.com/app/b" {
		t.Fatalf("first package = %q, want example.com/app/b", report.Packages[0].Package)
	}
	if report.Packages[0].ProductionEffective.Delta != 40 {
		t.Fatalf("production delta = %d, want 40", report.Packages[0].ProductionEffective.Delta)
	}
	if report.Packages[1].ProductionEffective.Delta != -25 {
		t.Fatalf("production delta = %d, want -25", report.Packages[1].ProductionEffective.Delta)
	}
}
