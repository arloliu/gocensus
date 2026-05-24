package cli

import (
	"bytes"
	"testing"

	"github.com/arloliu/gocensus/internal/contrib"
)

func TestRenderWhoTableAlignsWideNumbers(t *testing.T) {
	report := contrib.Report{
		Root:  "/repo",
		Scope: "human-authored Go files (*.go, generated and mock paths excluded)",
		Notes: []string{
			"Feature, fix, and refactor counts are commit-message heuristics.",
			"Line, file, commit, and active-day counts come from git log --numstat.",
		},
		Contributors: []contrib.Contributor{
			{
				Name:       "ASCII_AUTHOR",
				Commits:    622,
				Features:   360,
				Fixes:      104,
				Refactors:  134,
				Added:      320742,
				Removed:    133589,
				Net:        187153,
				Churn:      454331,
				Files:      2615,
				ActiveDays: 236,
			},
			{
				Name:       "wide_label",
				Commits:    189,
				Features:   104,
				Fixes:      23,
				Refactors:  47,
				Added:      51790,
				Removed:    53508,
				Net:        -1718,
				Churn:      105298,
				Files:      728,
				ActiveDays: 83,
			},
		},
	}
	var out bytes.Buffer

	if err := renderWhoTable(&out, report, "commits"); err != nil {
		t.Fatal(err)
	}

	want := "" +
		"Who: /repo\n" +
		"Scope: human-authored Go files (*.go, generated and mock paths excluded)\n" +
		"Sorted by: commits\n\n" +
		"  Author        Commits  Feat  Fix  Refactor    Added  Removed      Net    Churn  Files  Active Days\n" +
		"  ASCII_AUTHOR      622   360  104       134  320,742  133,589  187,153  454,331  2,615          236\n" +
		"  wide_label        189   104   23        47   51,790   53,508   -1,718  105,298    728           83\n\n" +
		"Notes:\n" +
		"  Feature/Fix/Refactor: commit-message heuristics.\n" +
		"  Line metrics: git log --numstat.\n"
	if out.String() != want {
		t.Fatalf("renderWhoTable() =\n%s\nwant:\n%s", out.String(), want)
	}
}
