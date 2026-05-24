package hotspot_test

import (
	"strings"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/hotspot"
)

func TestRankFiltersToProductionFilesAndScoresWithChurn(t *testing.T) {
	result := gocensus.Result{
		Root:  "/repo",
		Scope: "production excludes generated and mock files",
		FileMetrics: []gocensus.FileMetric{
			{Path: "main.go", Package: "example.com/app", Kind: "production", CodeLines: 100},
			{Path: "main_test.go", Package: "example.com/app", Kind: "test", CodeLines: 200},
			{Path: "service.pb.go", Package: "example.com/app", Kind: "generated", CodeLines: 300},
		},
		Packages: []gocensus.PackageMetric{
			{
				ImportPath: "example.com/app",
				Lines: gocensus.LineCounts{
					Production: gocensus.Metric{Effective: 100},
					Tests:      gocensus.Metric{Effective: 50},
				},
				Ratios: gocensus.Ratios{TestToProductionEffective: 0.5},
			},
		},
	}
	churn := map[string]hotspot.Churn{
		"main.go": {Added: 70, Removed: 30, Commits: 3},
	}

	report, err := hotspot.Rank(result, churn, hotspot.Options{By: "score", Top: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 {
		t.Fatalf("file count = %d, want 1", len(report.Files))
	}
	got := report.Files[0]
	if got.Path != "main.go" {
		t.Fatalf("path = %q, want main.go", got.Path)
	}
	if got.Score != 200 {
		t.Fatalf("score = %d, want 200", got.Score)
	}
	if got.Churn != 100 {
		t.Fatalf("churn = %d, want 100", got.Churn)
	}
	if got.PackageTestRatio != 0.5 {
		t.Fatalf("package test ratio = %v, want 0.5", got.PackageTestRatio)
	}
}

func TestRankCanIncludeGeneratedAndMockProductionScopeFiles(t *testing.T) {
	result := gocensus.Result{
		Root:  "/repo",
		Scope: "production includes generated and mock files",
		FileMetrics: []gocensus.FileMetric{
			{Path: "main.go", Package: "example.com/app", Kind: "production", CodeLines: 100},
			{Path: "service.pb.go", Package: "example.com/app", Kind: "generated", CodeLines: 300},
			{Path: "mock_client.go", Package: "example.com/app", Kind: "mock", CodeLines: 200},
		},
	}

	report, err := hotspot.Rank(result, nil, hotspot.Options{
		By:               "lines",
		Top:              0,
		IncludeGenerated: true,
		IncludeMocks:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := paths(report.Files); strings.Join(got, ",") != "service.pb.go,mock_client.go,main.go" {
		t.Fatalf("paths = %v, want generated, mock, production files", got)
	}
}

func TestParseNumstatAggregatesChurnByPath(t *testing.T) {
	input := strings.NewReader("\x1eabc\x1f2026-01-03\n10\t2\tmain.go\n-\t-\tasset.bin\n\x1edef\x1f2026-01-04\n3\t4\tmain.go\n5\t0\tpkg/service.go\n")

	churn, err := hotspot.ParseNumstat(input)
	if err != nil {
		t.Fatal(err)
	}
	if churn["main.go"].Added != 13 || churn["main.go"].Removed != 6 || churn["main.go"].Commits != 2 {
		t.Fatalf("main.go churn = %#v", churn["main.go"])
	}
	if churn["pkg/service.go"].Added != 5 || churn["pkg/service.go"].Removed != 0 || churn["pkg/service.go"].Commits != 1 {
		t.Fatalf("pkg/service.go churn = %#v", churn["pkg/service.go"])
	}
	if _, ok := churn["asset.bin"]; ok {
		t.Fatalf("binary file should be skipped: %#v", churn["asset.bin"])
	}
}

func paths(files []hotspot.File) []string {
	out := make([]string, 0, len(files))
	for _, file := range files {
		out = append(out, file.Path)
	}
	return out
}
