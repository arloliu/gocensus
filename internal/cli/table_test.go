package cli

import (
	"bytes"
	"testing"

	"github.com/arloliu/gocensus/internal/color"
)

func TestRenderTextTableAlignsWideRunes(t *testing.T) {
	table := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: "Author", Align: tableAlignLeft},
			{Header: "Commits", Align: tableAlignRight},
			{Header: "Net", Align: tableAlignRight},
		},
		Rows: [][]string{
			{"ASCII_AUTHOR", "569", "100,309"},
			{"寬字標籤", "18", "417"},
		},
	}
	var out bytes.Buffer

	if err := renderTextTable(&out, table); err != nil {
		t.Fatal(err)
	}

	want := "" +
		"  Author        Commits      Net\n" +
		"  ASCII_AUTHOR      569  100,309\n" +
		"  寬字標籤           18      417\n"
	if out.String() != want {
		t.Fatalf("renderTextTable() =\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRenderTextTableAlignsColoredCells(t *testing.T) {
	style := color.RGB()
	table := textTable{
		Indent: "  ",
		Gap:    "  ",
		Columns: []tableColumn{
			{Header: style.Header("Author"), Align: tableAlignLeft},
			{Header: style.Header("Commits"), Align: tableAlignRight},
			{Header: style.Header("Net"), Align: tableAlignRight},
		},
		Rows: [][]string{
			{style.Label("ASCII_AUTHOR"), style.Metric("569"), style.Metric("100,309")},
			{style.Label("Bob"), style.Metric("18"), style.Bad("-417")},
		},
	}
	var out bytes.Buffer

	if err := renderTextTable(&out, table); err != nil {
		t.Fatal(err)
	}

	got := color.StripSGR(out.String())
	want := "" +
		"  Author        Commits      Net\n" +
		"  ASCII_AUTHOR      569  100,309\n" +
		"  Bob                18     -417\n"
	if got != want {
		t.Fatalf("visible table =\n%s\nwant:\n%s", got, want)
	}
}
