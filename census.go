package gocensus

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/arloliu/gocensus/internal/classify"
	"github.com/arloliu/gocensus/internal/count"
	"github.com/arloliu/gocensus/internal/discover"
)

// Options controls repository analysis.
type Options struct {
	Root             string
	NoGitignore      bool
	ExtraExcludes    []string
	IncludeGenerated bool
	IncludeMocks     bool
}

// Result contains repository census metrics.
type Result struct {
	Root        string          `json:"root"`
	ModulePath  string          `json:"module_path"`
	Files       FileCounts      `json:"files"`
	Lines       LineCounts      `json:"lines"`
	Tests       TestCounts      `json:"tests"`
	Ratios      Ratios          `json:"ratios"`
	Packages    []PackageMetric `json:"packages"`
	FileMetrics []FileMetric    `json:"file_metrics"`
}

// FileCounts summarizes files by role.
type FileCounts struct {
	Total      int `json:"total"`
	Production int `json:"production"`
	Tests      int `json:"tests"`
	Generated  int `json:"generated"`
	Mocks      int `json:"mocks"`
}

// LineCounts summarizes raw and effective source lines.
type LineCounts struct {
	Production Metric `json:"production"`
	Tests      Metric `json:"tests"`
	Generated  Metric `json:"generated"`
	Mocks      Metric `json:"mocks"`
}

// Metric stores raw and effective line counts.
type Metric struct {
	Raw       int `json:"raw"`
	Effective int `json:"effective"`
}

// TestCounts summarizes Go test declarations.
type TestCounts struct {
	Tests      int `json:"tests"`
	Benchmarks int `json:"benchmarks"`
	Examples   int `json:"examples"`
}

// Ratios summarizes derived repository or package ratios.
type Ratios struct {
	TestToProductionRaw       float64 `json:"test_to_production_raw"`
	TestToProductionEffective float64 `json:"test_to_production_effective"`
	TestShareEffective        float64 `json:"test_share_effective"`
	GeneratedShareRaw         float64 `json:"generated_share_raw"`
	MockShareRaw              float64 `json:"mock_share_raw"`
}

// PackageMetric contains census metrics for one Go package.
type PackageMetric struct {
	ImportPath string     `json:"import_path"`
	Dir        string     `json:"dir"`
	Files      FileCounts `json:"files"`
	Lines      LineCounts `json:"lines"`
	Tests      TestCounts `json:"tests"`
	Ratios     Ratios     `json:"ratios"`
}

// FileMetric contains census metrics for one Go source file.
type FileMetric struct {
	Path       string `json:"path"`
	Package    string `json:"package"`
	Kind       string `json:"kind"`
	Generated  bool   `json:"generated"`
	RawLines   int    `json:"raw_lines"`
	CodeLines  int    `json:"code_lines"`
	Tests      int    `json:"tests"`
	Benchmarks int    `json:"benchmarks"`
	Examples   int    `json:"examples"`
}

// Analyze returns census metrics for a Go repository.
func Analyze(ctx context.Context, opts Options) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	root := opts.Root
	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return Result{}, err
	}
	if !info.IsDir() {
		return Result{}, errors.New("root is not a directory")
	}

	modulePath, err := readModulePath(filepath.Join(absRoot, "go.mod"))
	if err != nil {
		return Result{}, err
	}

	files, err := discover.GoFiles(ctx, discover.Options{
		Root:          absRoot,
		UseGitignore:  !opts.NoGitignore,
		ExtraExcludes: opts.ExtraExcludes,
	})
	if err != nil {
		return Result{}, err
	}

	fileMetrics := make([]FileMetric, 0, len(files))
	for _, path := range files {
		kind, err := classify.File(path)
		if err != nil {
			return Result{}, err
		}
		metrics, err := count.File(path)
		if err != nil {
			return Result{}, err
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return Result{}, err
		}

		kindString := string(kind)
		if kind == classify.KindGenerated && opts.IncludeGenerated {
			kindString = string(classify.KindProduction)
		}
		if kind == classify.KindMock && opts.IncludeMocks {
			kindString = string(classify.KindProduction)
		}
		fileMetrics = append(fileMetrics, FileMetric{
			Path:       filepath.ToSlash(rel),
			Package:    metrics.Package,
			Kind:       kindString,
			Generated:  kind == classify.KindGenerated,
			RawLines:   metrics.RawLines,
			CodeLines:  metrics.CodeLines,
			Tests:      metrics.Tests,
			Benchmarks: metrics.Benchmarks,
			Examples:   metrics.Examples,
		})
	}

	return buildResult(absRoot, modulePath, fileMetrics), nil
}

func buildResult(root string, modulePath string, fileMetrics []FileMetric) Result {
	result := Result{
		Root:        root,
		ModulePath:  modulePath,
		FileMetrics: append([]FileMetric(nil), fileMetrics...),
	}

	byPackage := map[string]*PackageMetric{}
	for _, file := range fileMetrics {
		addFile(&result.Files, &result.Lines, &result.Tests, file)
		pkg := byPackage[file.Package]
		if pkg == nil {
			pkg = &PackageMetric{
				ImportPath: file.Package,
				Dir:        file.Package,
			}
			byPackage[file.Package] = pkg
		}
		addFile(&pkg.Files, &pkg.Lines, &pkg.Tests, file)
	}
	result.Ratios = ratios(result.Lines)

	for _, pkg := range byPackage {
		pkg.Ratios = ratios(pkg.Lines)
		result.Packages = append(result.Packages, *pkg)
	}
	slices.SortFunc(result.Packages, func(a, b PackageMetric) int {
		if a.ImportPath < b.ImportPath {
			return -1
		}
		if a.ImportPath > b.ImportPath {
			return 1
		}
		return 0
	})
	return result
}

func addFile(files *FileCounts, lines *LineCounts, tests *TestCounts, file FileMetric) {
	files.Total++
	tests.Tests += file.Tests
	tests.Benchmarks += file.Benchmarks
	tests.Examples += file.Examples

	metric := Metric{Raw: file.RawLines, Effective: file.CodeLines}
	switch file.Kind {
	case "production":
		files.Production++
		lines.Production.Raw += metric.Raw
		lines.Production.Effective += metric.Effective
	case "test":
		files.Tests++
		lines.Tests.Raw += metric.Raw
		lines.Tests.Effective += metric.Effective
	case "generated":
		files.Generated++
		lines.Generated.Raw += metric.Raw
		lines.Generated.Effective += metric.Effective
	case "mock":
		files.Mocks++
		lines.Mocks.Raw += metric.Raw
		lines.Mocks.Effective += metric.Effective
	}
}

func ratios(lines LineCounts) Ratios {
	totalRaw := lines.Production.Raw + lines.Tests.Raw + lines.Generated.Raw + lines.Mocks.Raw
	totalEffective := lines.Production.Effective + lines.Tests.Effective
	return Ratios{
		TestToProductionRaw:       divide(lines.Tests.Raw, lines.Production.Raw),
		TestToProductionEffective: divide(lines.Tests.Effective, lines.Production.Effective),
		TestShareEffective:        divide(lines.Tests.Effective, totalEffective),
		GeneratedShareRaw:         divide(lines.Generated.Raw, totalRaw),
		MockShareRaw:              divide(lines.Mocks.Raw, totalRaw),
	}
}

func divide(numerator int, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func readModulePath(path string) (string, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if modulePath, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(modulePath), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}
