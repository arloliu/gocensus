package gocensus_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus"
)

func TestAnalyzeEmptyGoModule(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := gocensus.Analyze(context.Background(), gocensus.Options{Root: dir})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.Root != dir {
		t.Fatalf("Root = %q, want %q", result.Root, dir)
	}
	if result.ModulePath != "example.com/app" {
		t.Fatalf("ModulePath = %q, want example.com/app", result.ModulePath)
	}
	if result.Files.Total != 0 {
		t.Fatalf("Files.Total = %d, want 0", result.Files.Total)
	}
	if len(result.Packages) != 0 {
		t.Fatalf("Packages length = %d, want 0", len(result.Packages))
	}
	if len(result.FileMetrics) != 0 {
		t.Fatalf("FileMetrics length = %d, want 0", len(result.FileMetrics))
	}
}
