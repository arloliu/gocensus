package count

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Metrics contains source metrics for one Go file.
type Metrics struct {
	RawLines                 int
	CodeLines                int
	Tests                    int
	StaticSubtests           int
	DynamicSubtestSites      int
	Benchmarks               int
	StaticSubbenchmarks      int
	DynamicSubbenchmarkSites int
	Examples                 int
	Package                  string
}

// File counts raw lines, effective code lines, and test declarations.
func File(path string) (Metrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Metrics{}, err
	}

	metrics := Metrics{
		RawLines:  rawLines(content),
		CodeLines: effectiveLines(content),
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		return metrics, nil
	}
	metrics.Package = file.Name.Name
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		switch {
		case strings.HasPrefix(fn.Name.Name, "Test"):
			metrics.Tests++
			receiver := firstParamName(fn)
			stats := runStatsFor(fn, receiver)
			metrics.StaticSubtests += stats.staticRuns
			metrics.DynamicSubtestSites += stats.dynamicRunSites
		case strings.HasPrefix(fn.Name.Name, "Benchmark"):
			metrics.Benchmarks++
			receiver := firstParamName(fn)
			stats := runStatsFor(fn, receiver)
			metrics.StaticSubbenchmarks += stats.staticRuns
			metrics.DynamicSubbenchmarkSites += stats.dynamicRunSites
		case strings.HasPrefix(fn.Name.Name, "Example"):
			metrics.Examples++
		}
	}
	return metrics, nil
}
