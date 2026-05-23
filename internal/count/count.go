package count

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"strings"
)

// Metrics contains source metrics for one Go file.
type Metrics struct {
	RawLines   int
	CodeLines  int
	Tests      int
	Benchmarks int
	Examples   int
	Package    string
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
		case strings.HasPrefix(fn.Name.Name, "Benchmark"):
			metrics.Benchmarks++
		case strings.HasPrefix(fn.Name.Name, "Example"):
			metrics.Examples++
		}
	}
	return metrics, nil
}

func rawLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}

func effectiveLines(content []byte) int {
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(content))
	var scan scanner.Scanner
	scan.Init(file, content, nil, 0)

	lines := map[int]struct{}{}
	for {
		pos, tok, _ := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			continue
		}
		lines[fileSet.Position(pos).Line] = struct{}{}
	}
	return len(lines)
}
