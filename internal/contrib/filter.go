package contrib

import (
	"path/filepath"
	"strings"
)

func keepPath(path string, opts ParseOptions) bool {
	if opts.GoOnly && !strings.HasSuffix(strings.ToLower(path), ".go") {
		return false
	}
	if excludeGenerated(opts) && isGeneratedPath(path) {
		return false
	}
	if excludeMocks(opts) && isMockPath(path) {
		return false
	}
	if excludeTestdata(opts) && isTestdataPath(path) {
		return false
	}
	return true
}

func excludeGenerated(opts ParseOptions) bool {
	return !opts.IncludeGenerated
}

func excludeMocks(opts ParseOptions) bool {
	return !opts.IncludeMocks
}

func excludeTestdata(opts ParseOptions) bool {
	return !opts.IncludeTestdata
}

func isGeneratedPath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lower)
	if strings.HasSuffix(base, ".pb.go") ||
		strings.HasSuffix(base, ".pb.gw.go") ||
		strings.HasSuffix(base, "_gen.go") ||
		strings.HasSuffix(base, ".gen.go") ||
		strings.HasPrefix(base, "generated") {
		return true
	}
	for _, part := range strings.Split(lower, "/") {
		if part == "gen" || part == "generated" {
			return true
		}
	}
	return false
}

func isMockPath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lower)
	if strings.Contains(base, "mock") {
		return true
	}
	for _, part := range strings.Split(lower, "/") {
		if part == "mock" || part == "mocks" {
			return true
		}
	}
	return false
}

func isTestdataPath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	for _, part := range strings.Split(lower, "/") {
		if part == "testdata" {
			return true
		}
	}
	return false
}
