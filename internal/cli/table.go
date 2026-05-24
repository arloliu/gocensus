package cli

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/arloliu/gocensus/internal/color"
)

type tableAlign int

const (
	tableAlignLeft tableAlign = iota
	tableAlignRight
)

type tableColumn struct {
	Header string
	Align  tableAlign
}

type textTable struct {
	Columns []tableColumn
	Rows    [][]string
	Indent  string
	Gap     string
}

func renderTextTable(w io.Writer, table textTable) error {
	widths := tableColumnWidths(table)
	if err := renderTextTableRow(w, table.Indent, table.Gap, tableHeaderRow(table.Columns), widths, table.Columns); err != nil {
		return err
	}
	for _, row := range table.Rows {
		if err := renderTextTableRow(w, table.Indent, table.Gap, row, widths, table.Columns); err != nil {
			return err
		}
	}
	return nil
}

func tableHeaderRow(columns []tableColumn) []string {
	row := make([]string, 0, len(columns))
	for _, column := range columns {
		row = append(row, column.Header)
	}
	return row
}

func tableColumnWidths(table textTable) []int {
	widths := make([]int, len(table.Columns))
	for i, column := range table.Columns {
		widths[i] = displayWidth(column.Header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if width := displayWidth(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}
	return widths
}

func renderTextTableRow(w io.Writer, indent string, gap string, row []string, widths []int, columns []tableColumn) error {
	if _, err := fmt.Fprint(w, indent); err != nil {
		return err
	}
	for i, column := range columns {
		if i > 0 {
			if _, err := fmt.Fprint(w, gap); err != nil {
				return err
			}
		}
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		padded := padDisplay(cell, widths[i], column.Align)
		if _, err := fmt.Fprint(w, padded); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func padDisplay(value string, width int, align tableAlign) string {
	padding := width - displayWidth(value)
	if padding <= 0 {
		return value
	}
	spaces := strings.Repeat(" ", padding)
	if align == tableAlignRight {
		return spaces + value
	}
	return value + spaces
}

func displayWidth(value string) int {
	value = color.StripSGR(value)
	width := 0
	for _, r := range value {
		switch {
		case r == '\t':
			width += 4
		case r == '\n' || r == '\r':
		case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r):
		case isWideRune(r):
			width += 2
		default:
			width++
		}
	}
	return width
}

func isWideRune(r rune) bool {
	return (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2329 && r <= 0x232A) ||
		(r >= 0x2E80 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6)
}
