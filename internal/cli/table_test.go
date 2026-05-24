package cli

import (
	"bytes"
	"testing"
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
