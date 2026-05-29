package render_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/color"
	"github.com/arloliu/gocensus/internal/render"
)

func TestTableIncludesCoreSections(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "table")
	if err != nil {
		t.Fatal(err)
	}
	want := `Go Census: example.com/app
Scope: production excludes generated and mock files; testdata directories excluded

Overview
  Go files: 15    Packages: 1    Known test cases: 59

Code Mix
  Kind                 Files   Raw Lines   Effective Lines
  Production Scope        8       1,192             1,012
  Tests                   7         528               448
  Excluded Generated      2         200               160
  Excluded Mocks          1          30                20
  Total                  15       1,720             1,460

Ratios
  Test / Production Scope       0.44:1
  Test Share                     30.7%

Test Inventory
  Known Test Cases        59
  Top-level Tests         17
  Static Subtests         42
  Dynamic Subtest Sites    3
  Known Benchmark Cases    6
  Benchmarks               2
  Static Subbenchmarks     4
  Dynamic Benchmark Sites  1
  Examples                 1

Notes
  Raw Lines            Physical lines, including blanks and comments.
  Effective Lines      Lines containing non-comment Go tokens.
  Production Scope     Non-test Go files counted as production; scope line shows generated/mock/testdata inclusion.
  Tests                *_test.go files.
  Known Cases          Top-level tests plus statically countable subtests.
  Static Subtests      t.Run/b.Run cases with statically countable case data.
  Dynamic Sites        t.Run/b.Run call sites with runtime-dependent case counts.
  Excluded Generated   Generated files not counted in production scope.
  Excluded Mocks       Mock/support files not counted in production scope.
`
	if out.String() != want {
		t.Fatalf("table output mismatch\nwant:\n%s\ngot:\n%s", want, out.String())
	}
}

func TestTablePlainOutputContainsNoSGR(t *testing.T) {
	var out bytes.Buffer
	err := render.Result(&out, sample(), "table")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("plain table output contains SGR: %q", out.String())
	}
}

func TestTableCanRenderRGBColor(t *testing.T) {
	var out bytes.Buffer
	err := render.ResultWithOptions(&out, sample(), "table", render.Options{Style: color.RGB()})
	if err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "\x1b[38;2;") {
		t.Fatalf("colored table output missing RGB SGR: %q", text)
	}
	if !strings.Contains(text, "Go Census") {
		t.Fatalf("colored table output missing title: %q", text)
	}
}

func TestJSONIgnoresColorOptions(t *testing.T) {
	var out bytes.Buffer
	err := render.ResultWithOptions(&out, sample(), "json", render.Options{Style: color.RGB()})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("json output contains SGR: %q", out.String())
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
	if !strings.Contains(text, "Scope: production excludes generated and mock files; testdata directories excluded") {
		t.Fatalf("markdown missing scope:\n%s", text)
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
		Scope:      "production excludes generated and mock files; testdata directories excluded",
		Files:      gocensus.FileCounts{Total: 15, Production: 8, Tests: 7, Generated: 2, Mocks: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Raw: 1192, Effective: 1012},
			Tests:      gocensus.Metric{Raw: 528, Effective: 448},
			Generated:  gocensus.Metric{Raw: 200, Effective: 160},
			Mocks:      gocensus.Metric{Raw: 30, Effective: 20},
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
