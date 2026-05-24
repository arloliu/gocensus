package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/arloliu/gocensus/internal/contrib"
)

func renderWho(w io.Writer, report contrib.Report, sortBy string, format string, output string) error {
	var out bytes.Buffer
	writer := w
	if output != "" {
		writer = &out
	}
	switch format {
	case "table":
		if err := renderWhoTable(writer, report, sortBy); err != nil {
			return err
		}
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return err
		}
	case "markdown":
		if err := renderWhoMarkdown(writer, report, sortBy); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if output != "" {
		if err := os.WriteFile(output, out.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func renderWhoTable(w io.Writer, report contrib.Report, sortBy string) error {
	if _, err := fmt.Fprintf(w, "Who: %s\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: %s\n\n", sortBy); err != nil {
		return err
	}

	table := textTable{
		Indent:  "  ",
		Gap:     "  ",
		Columns: whoTableColumns(),
		Rows:    whoRows(report.Contributors),
	}
	if err := renderTextTable(w, table); err != nil {
		return err
	}
	if len(report.Notes) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Notes:"); err != nil {
			return err
		}
		for _, note := range report.Notes {
			if _, err := fmt.Fprintf(w, "  %s\n", displayWhoNote(note)); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderWhoMarkdown(w io.Writer, report contrib.Report, sortBy string) error {
	if _, err := fmt.Fprintf(w, "# Who: %s\n\n", report.Root); err != nil {
		return err
	}
	if report.Scope != "" {
		if _, err := fmt.Fprintf(w, "Scope: %s\n\n", report.Scope); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Sorted by: `%s`\n\n", sortBy); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Author | Commits | Feat | Fix | Refactor | Added | Removed | Net | Churn | Files | Active Days |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"); err != nil {
		return err
	}
	for _, contributor := range report.Contributors {
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			contributor.Name,
			formatWhoInt(contributor.Commits),
			formatWhoInt(contributor.Features),
			formatWhoInt(contributor.Fixes),
			formatWhoInt(contributor.Refactors),
			formatWhoInt(contributor.Added),
			formatWhoInt(contributor.Removed),
			formatWhoInt(contributor.Net),
			formatWhoInt(contributor.Churn),
			formatWhoInt(contributor.Files),
			formatWhoInt(contributor.ActiveDays),
		); err != nil {
			return err
		}
	}
	if len(report.Notes) > 0 {
		if _, err := fmt.Fprintln(w, "\n## Notes"); err != nil {
			return err
		}
		for _, note := range report.Notes {
			if _, err := fmt.Fprintf(w, "- %s\n", displayWhoNote(note)); err != nil {
				return err
			}
		}
	}
	return nil
}

func whoTableColumns() []tableColumn {
	return []tableColumn{
		{Header: "Author", Align: tableAlignLeft},
		{Header: "Commits", Align: tableAlignRight},
		{Header: "Feat", Align: tableAlignRight},
		{Header: "Fix", Align: tableAlignRight},
		{Header: "Refactor", Align: tableAlignRight},
		{Header: "Added", Align: tableAlignRight},
		{Header: "Removed", Align: tableAlignRight},
		{Header: "Net", Align: tableAlignRight},
		{Header: "Churn", Align: tableAlignRight},
		{Header: "Files", Align: tableAlignRight},
		{Header: "Active Days", Align: tableAlignRight},
	}
}

func whoRows(contributors []contrib.Contributor) [][]string {
	rows := make([][]string, 0, len(contributors))
	for _, contributor := range contributors {
		rows = append(rows, []string{
			contributor.Name,
			formatWhoInt(contributor.Commits),
			formatWhoInt(contributor.Features),
			formatWhoInt(contributor.Fixes),
			formatWhoInt(contributor.Refactors),
			formatWhoInt(contributor.Added),
			formatWhoInt(contributor.Removed),
			formatWhoInt(contributor.Net),
			formatWhoInt(contributor.Churn),
			formatWhoInt(contributor.Files),
			formatWhoInt(contributor.ActiveDays),
		})
	}
	return rows
}

func formatWhoInt(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	text := strconv.Itoa(value)
	if len(text) <= 3 {
		return sign + text
	}

	firstGroup := len(text) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}

	var out strings.Builder
	out.WriteString(text[:firstGroup])
	for i := firstGroup; i < len(text); i += 3 {
		out.WriteString("," + text[i:i+3])
	}
	return sign + out.String()
}

func displayWhoNote(note string) string {
	switch note {
	case "Feature, fix, and refactor counts are commit-message heuristics.":
		return "Feature/Fix/Refactor: commit-message heuristics."
	case "Line, file, commit, and active-day counts come from git log --numstat.":
		return "Line metrics: git log --numstat."
	default:
		return note
	}
}
