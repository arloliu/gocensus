package render_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/render"
)

func TestTableIncludesCoreSections(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "table")
	if err != nil {
		t.Fatal(err)
	}
	want := `Go Census: example.com/app

Overview
  Go files: 15    Packages: 1    Test funcs: 17

Code Mix
  Kind          Files   Raw Lines   Effective Lines
  Production       8       1,192             1,012
  Tests            7         528               448
  Generated        0           0                 0
  Mocks            0           0                 0
  Total           15       1,720             1,460

Ratios
  Test / Production       0.44:1
  Test Share               30.7%
  Generated Share           0.0%
  Mock Share                0.0%

Test Inventory
  Tests                   17
  Static Subtests         42
  Dynamic Subtest Sites    3
  Benchmarks               2
  Static Subbenchmarks     4
  Dynamic Benchmark Sites  1
  Examples                 1

Notes
  Raw Lines        Physical lines, including blanks and comments.
  Effective Lines  Lines containing non-comment Go tokens.
  Production       Non-test Go files, excluding generated files and mocks.
  Tests            *_test.go files.
  Static Subtests  t.Run/b.Run cases with statically countable case data.
  Dynamic Sites    t.Run/b.Run call sites with runtime-dependent case counts.
  Generated        Files with generated-code markers or generated suffixes.
  Mocks            Files classified as mock/support code.
`
	if out.String() != want {
		t.Fatalf("table output mismatch\nwant:\n%s\ngot:\n%s", want, out.String())
	}
}

func TestMarkdownIncludesPackageTable(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "markdown")
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "| Package | Prod Lines | Test Lines | Test Ratio |") {
		t.Fatalf("markdown missing package table:\n%s", text)
	}
}

func TestJSONUsesStableFieldNames(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "json")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["module_path"]; !ok {
		t.Fatalf("json missing module_path: %s", out.String())
	}
	if _, ok := payload["file_metrics"]; !ok {
		t.Fatalf("json missing file_metrics: %s", out.String())
	}
}

func sample() gocensus.Result {
	return gocensus.Result{
		Root:       "/repo",
		ModulePath: "example.com/app",
		Files:      gocensus.FileCounts{Total: 15, Production: 8, Tests: 7},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Raw: 1192, Effective: 1012},
			Tests:      gocensus.Metric{Raw: 528, Effective: 448},
		},
		Tests: gocensus.TestCounts{
			Tests:                    17,
			StaticSubtests:           42,
			DynamicSubtestSites:      3,
			Benchmarks:               2,
			StaticSubbenchmarks:      4,
			DynamicSubbenchmarkSites: 1,
			Examples:                 1,
		},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.442687747, TestShareEffective: 0.306849315},
		Packages: []gocensus.PackageMetric{{
			ImportPath: "main",
			Lines: gocensus.LineCounts{
				Production: gocensus.Metric{Raw: 10, Effective: 8},
				Tests:      gocensus.Metric{Raw: 6, Effective: 4},
			},
			Ratios: gocensus.Ratios{TestToProductionEffective: 0.5},
		}},
	}
}
