package hotspot

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/arloliu/gocensus"
)

const (
	// SortScore orders files by effective lines plus Git churn.
	SortScore = "score"
	// SortLines orders files by effective lines.
	SortLines = "lines"
	// SortChurn orders files by added plus removed lines.
	SortChurn = "churn"
	// SortCommits orders files by commit count.
	SortCommits = "commits"
	// SortRatio orders files by package test-to-production ratio.
	SortRatio = "test-ratio"
)

// Options controls hotspot ranking.
type Options struct {
	By               string
	Top              int
	IncludeGenerated bool
	IncludeMocks     bool
}

// Churn contains Git numstat metrics for a path.
type Churn struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Commits int `json:"commits"`
}

// Report contains ranked file hotspots.
type Report struct {
	Root   string   `json:"root"`
	Scope  string   `json:"scope"`
	SortBy string   `json:"sort_by"`
	Notes  []string `json:"notes,omitempty"`
	Files  []File   `json:"files"`
}

// File contains one file hotspot row.
type File struct {
	Path             string  `json:"path"`
	Package          string  `json:"package"`
	Score            int     `json:"score"`
	EffectiveLines   int     `json:"effective_lines"`
	Churn            int     `json:"churn"`
	Added            int     `json:"added"`
	Removed          int     `json:"removed"`
	Commits          int     `json:"commits"`
	PackageTestRatio float64 `json:"package_test_ratio"`
}

// Rank combines current census metrics with Git churn and returns hotspot rows.
func Rank(result gocensus.Result, churnByPath map[string]Churn, opts Options) (Report, error) {
	if opts.By == "" {
		opts.By = SortScore
	}
	if !validSort(opts.By) {
		return Report{}, fmt.Errorf("unknown hotspot sort %q", opts.By)
	}

	ratios := packageRatios(result.Packages)
	files := make([]File, 0, len(result.FileMetrics))
	for _, metric := range result.FileMetrics {
		if !includeFile(metric, opts) {
			continue
		}
		churn := churnByPath[metric.Path]
		totalChurn := churn.Added + churn.Removed
		files = append(files, File{
			Path:             metric.Path,
			Package:          metric.Package,
			Score:            metric.CodeLines + totalChurn,
			EffectiveLines:   metric.CodeLines,
			Churn:            totalChurn,
			Added:            churn.Added,
			Removed:          churn.Removed,
			Commits:          churn.Commits,
			PackageTestRatio: ratios[metric.Package],
		})
	}
	sortFiles(files, opts.By)
	if opts.Top > 0 && len(files) > opts.Top {
		files = files[:opts.Top]
	}

	return Report{
		Root:   result.Root,
		Scope:  result.Scope,
		SortBy: opts.By,
		Notes: []string{
			"Hotspot Score = effective production lines + Git churn.",
			"Git Churn = added + removed lines from git log --numstat.",
			"Package Test Ratio = package test effective lines divided by package production effective lines.",
		},
		Files: files,
	}, nil
}

func includeFile(metric gocensus.FileMetric, opts Options) bool {
	switch metric.Kind {
	case "production":
		return true
	case "generated":
		return opts.IncludeGenerated
	case "mock":
		return opts.IncludeMocks
	default:
		return false
	}
}

// ParseNumstat aggregates git log --numstat output by path.
func ParseNumstat(r io.Reader) (map[string]Churn, error) {
	scanner := bufio.NewScanner(r)
	out := map[string]Churn{}
	seenInCommit := map[string]struct{}{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "\x1e") {
			seenInCommit = map[string]struct{}{}
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 || fields[0] == "-" || fields[1] == "-" {
			continue
		}
		added, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, err
		}
		removed, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}

		path := filepath.ToSlash(fields[2])
		churn := out[path]
		churn.Added += added
		churn.Removed += removed
		if _, ok := seenInCommit[path]; !ok {
			churn.Commits++
			seenInCommit[path] = struct{}{}
		}
		out[path] = churn
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func validSort(sortBy string) bool {
	switch sortBy {
	case SortScore, SortLines, SortChurn, SortCommits, SortRatio:
		return true
	default:
		return false
	}
}

func packageRatios(packages []gocensus.PackageMetric) map[string]float64 {
	ratios := make(map[string]float64, len(packages))
	for _, pkg := range packages {
		ratios[pkg.ImportPath] = pkg.Ratios.TestToProductionEffective
	}
	return ratios
}

func sortFiles(files []File, sortBy string) {
	slices.SortFunc(files, func(a File, b File) int {
		switch sortBy {
		case SortScore:
			if a.Score != b.Score {
				return descending(a.Score, b.Score)
			}
		case SortLines:
			if a.EffectiveLines != b.EffectiveLines {
				return descending(a.EffectiveLines, b.EffectiveLines)
			}
		case SortChurn:
			if a.Churn != b.Churn {
				return descending(a.Churn, b.Churn)
			}
		case SortCommits:
			if a.Commits != b.Commits {
				return descending(a.Commits, b.Commits)
			}
		case SortRatio:
			if a.PackageTestRatio > b.PackageTestRatio {
				return -1
			}
			if a.PackageTestRatio < b.PackageTestRatio {
				return 1
			}
		}
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})
}

func descending(a int, b int) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}
