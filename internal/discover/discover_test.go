package discover_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/arloliu/gocensus/internal/discover"
)

func TestGoFilesRespectsDefaultExcludesAndGitignore(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".gitignore", "ignored.go\nnested/ignored_dir/\n")
	write(t, root, "main.go", "package main\n")
	write(t, root, "ignored.go", "package main\n")
	write(t, root, "vendor/lib/lib.go", "package lib\n")
	write(t, root, ".hidden/hidden.go", "package hidden\n")
	write(t, root, "testdata/fixture.go", "package fixture\n")
	write(t, root, "nested/keep.go", "package nested\n")
	write(t, root, "nested/ignored_dir/drop.go", "package ignored\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:         root,
		UseGitignore: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"main.go", "nested/keep.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func TestGoFilesCanDisableGitignore(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".gitignore", "ignored.go\n")
	write(t, root, "main.go", "package main\n")
	write(t, root, "ignored.go", "package main\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:         root,
		UseGitignore: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"ignored.go", "main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func TestGoFilesCanIncludeTestdata(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	write(t, root, "testdata/fixture.go", "package fixture\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:            root,
		UseGitignore:    true,
		IncludeTestdata: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"main.go", "testdata/fixture.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func TestGoFilesUsesExtraExcludes(t *testing.T) {
	root := t.TempDir()
	write(t, root, "main.go", "package main\n")
	write(t, root, "internal/drop.go", "package internal\n")

	files, err := discover.GoFiles(context.Background(), discover.Options{
		Root:          root,
		UseGitignore:  true,
		ExtraExcludes: []string{"internal/**"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got := rels(t, root, files)
	want := []string{"main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v, want %#v", got, want)
	}
}

func write(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func rels(t *testing.T, root string, files []string) []string {
	t.Helper()
	out := make([]string, 0, len(files))
	for _, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	slices.Sort(out)
	return out
}
