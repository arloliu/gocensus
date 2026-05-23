package classify

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Kind identifies the role of a Go source file in census metrics.
type Kind string

const (
	// KindProduction identifies non-test, non-generated, non-mock Go files.
	KindProduction Kind = "production"
	// KindTest identifies Go test files.
	KindTest Kind = "test"
	// KindGenerated identifies generated Go files.
	KindGenerated Kind = "generated"
	// KindMock identifies mock or mock-like Go files.
	KindMock Kind = "mock"
)

// File classifies a Go source file by path and leading generated-code marker.
func File(path string) (Kind, error) {
	base := filepath.Base(path)
	lower := strings.ToLower(base)

	if strings.HasSuffix(lower, ".pb.go") || hasGeneratedMarker(path) {
		return KindGenerated, nil
	}
	if strings.HasSuffix(lower, "_test.go") {
		return KindTest, nil
	}
	if strings.Contains(lower, "mock") {
		return KindMock, nil
	}
	return KindProduction, nil
}

func hasGeneratedMarker(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for i := 0; i < 20; i++ {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if strings.Contains(line, "Code generated") && strings.Contains(line, "DO NOT EDIT.") {
			return true
		}
	}
	return false
}
