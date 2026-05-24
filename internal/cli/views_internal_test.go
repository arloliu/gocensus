package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/color"
	censusdiff "github.com/arloliu/gocensus/internal/diff"
	"github.com/arloliu/gocensus/internal/hotspot"
)

func TestRenderDiffTableIsSelfExplaining(t *testing.T) {
	report := censusdiff.Report{
		Root:  "example.com/app",
		Base:  "base",
		Head:  "head",
		Scope: "production excludes generated and mock files",
		Summary: censusdiff.Summary{
			TotalFiles:          censusdiff.IntDelta{Base: 1, Head: 2, Delta: 1},
			ProductionFiles:     censusdiff.IntDelta{Base: 1, Head: 2, Delta: 1},
			TestFiles:           censusdiff.IntDelta{Base: 0, Head: 0, Delta: 0},
			ProductionEffective: censusdiff.IntDelta{Base: 10, Head: 20, Delta: 10},
			TestEffective:       censusdiff.IntDelta{Base: 5, Head: 5, Delta: 0},
			TestToProduction:    censusdiff.FloatDelta{Base: 0.5, Head: 0.25, Delta: -0.25},
			TestShare:           censusdiff.FloatDelta{Base: 0.3333333333, Head: 0.2, Delta: -0.1333333333},
		},
	}
	var out bytes.Buffer

	if err := renderDiff(&out, report, "table", "", color.Plain()); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Diff:", "Scope:", "Base:", "Head:", "Production Effective", "+10"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output = %q, want %q", out.String(), want)
		}
	}
}

func TestRenderHotspotsTableIsSelfExplaining(t *testing.T) {
	report := hotspot.Report{
		Root:   "example.com/app",
		Scope:  "human-authored production Go files (*.go, generated and mock paths excluded)",
		SortBy: hotspot.SortScore,
		Notes: []string{
			"Hotspot Score = effective production lines + Git churn.",
			"Git Churn = added + removed lines from git log --numstat.",
			"Package Test Ratio = package test effective lines divided by package production effective lines.",
		},
		Files: []hotspot.File{
			{
				Path:             "internal/app/app.go",
				Package:          "example.com/app/internal/app",
				Score:            120,
				EffectiveLines:   80,
				Churn:            40,
				Added:            30,
				Removed:          10,
				Commits:          3,
				PackageTestRatio: 0.75,
			},
		},
	}
	var out bytes.Buffer

	if err := renderHotspots(&out, report, "table", "", color.Plain()); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Hotspots:", "Scope:", "Hotspot Score", "Git Churn", "Package Test Ratio", "internal/app/app.go"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output = %q, want %q", out.String(), want)
		}
	}
}
