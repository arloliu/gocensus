package gocensus

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Options controls repository analysis.
type Options struct {
	Root string
}

// Result contains repository census metrics.
type Result struct {
	Root       string     `json:"root"`
	ModulePath string     `json:"module_path"`
	Files      FileCounts `json:"files"`
	Lines      LineCounts `json:"lines"`
	Tests      TestCounts `json:"tests"`
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

	return Result{
		Root:       absRoot,
		ModulePath: modulePath,
	}, nil
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
