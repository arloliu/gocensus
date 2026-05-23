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
	text := out.String()
	for _, want := range []string{"Go Census: example.com/app", "Files", "Lines", "Ratios", "Tests"} {
		if !strings.Contains(text, want) {
			t.Fatalf("table missing %q:\n%s", want, text)
		}
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
		Files:      gocensus.FileCounts{Total: 2, Production: 1, Tests: 1},
		Lines: gocensus.LineCounts{
			Production: gocensus.Metric{Raw: 10, Effective: 8},
			Tests:      gocensus.Metric{Raw: 6, Effective: 4},
		},
		Tests:  gocensus.TestCounts{Tests: 2},
		Ratios: gocensus.Ratios{TestToProductionEffective: 0.5, TestShareEffective: 0.3333333333},
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
