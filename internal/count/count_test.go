package count_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus/internal/count"
)

func TestFileCountsRawAndEffectiveLines(t *testing.T) {
	path := write(t, `package main

// package comment
func main() { // inline comment
	println("hi")
}

/*
block comment
*/
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.RawLines != 10 {
		t.Fatalf("RawLines = %d, want 10", metrics.RawLines)
	}
	if metrics.CodeLines != 4 {
		t.Fatalf("CodeLines = %d, want 4", metrics.CodeLines)
	}
}

func TestFileCountsTestDeclarations(t *testing.T) {
	path := write(t, `package main

import "testing"

func TestOne(t *testing.T) {}
func Testhelper(t *testing.T) {}
func BenchmarkOne(b *testing.B) {}
func ExampleOne() {}
func helper() {}
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.Tests != 2 {
		t.Fatalf("Tests = %d, want 2", metrics.Tests)
	}
	if metrics.Benchmarks != 1 {
		t.Fatalf("Benchmarks = %d, want 1", metrics.Benchmarks)
	}
	if metrics.Examples != 1 {
		t.Fatalf("Examples = %d, want 1", metrics.Examples)
	}
}

func write(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "file_test.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
