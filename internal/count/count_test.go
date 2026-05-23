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

func TestFileCountsStaticSubtests(t *testing.T) {
	path := write(t, `package main

import "testing"

func TestLiteralSubtests(t *testing.T) {
	t.Run("one", func(t *testing.T) {})
	t.Run("two", func(t *testing.T) {})
}

func BenchmarkLiteralSubbenchmarks(b *testing.B) {
	b.Run("one", func(b *testing.B) {})
}
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.StaticSubtests != 2 {
		t.Fatalf("StaticSubtests = %d, want 2", metrics.StaticSubtests)
	}
	if metrics.StaticSubbenchmarks != 1 {
		t.Fatalf("StaticSubbenchmarks = %d, want 1", metrics.StaticSubbenchmarks)
	}
}

func TestFileCountsTableDrivenSubtests(t *testing.T) {
	path := write(t, `package main

import "testing"

func TestTableSubtests(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "one"},
		{name: "two"},
		{name: "three"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {})
	}
}
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.StaticSubtests != 3 {
		t.Fatalf("StaticSubtests = %d, want 3", metrics.StaticSubtests)
	}
	if metrics.DynamicSubtestSites != 0 {
		t.Fatalf("DynamicSubtestSites = %d, want 0", metrics.DynamicSubtestSites)
	}
}

func TestFileCountsDynamicSubtestSites(t *testing.T) {
	path := write(t, `package main

import "testing"

func TestDynamicSubtests(t *testing.T) {
	for _, name := range loadCases() {
		t.Run(name, func(t *testing.T) {})
	}
}

func BenchmarkDynamicSubbenchmarks(b *testing.B) {
	for _, name := range loadCases() {
		b.Run(name, func(b *testing.B) {})
	}
}
`)

	metrics, err := count.File(path)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.StaticSubtests != 0 {
		t.Fatalf("StaticSubtests = %d, want 0", metrics.StaticSubtests)
	}
	if metrics.DynamicSubtestSites != 1 {
		t.Fatalf("DynamicSubtestSites = %d, want 1", metrics.DynamicSubtestSites)
	}
	if metrics.StaticSubbenchmarks != 0 {
		t.Fatalf("StaticSubbenchmarks = %d, want 0", metrics.StaticSubbenchmarks)
	}
	if metrics.DynamicSubbenchmarkSites != 1 {
		t.Fatalf("DynamicSubbenchmarkSites = %d, want 1", metrics.DynamicSubbenchmarkSites)
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
