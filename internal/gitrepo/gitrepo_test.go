package gitrepo_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/arloliu/gocensus/internal/gitrepo"
)

func TestRootReturnsTopLevelGitDirectory(t *testing.T) {
	dir := initRepo(t)
	nested := filepath.Join(dir, "internal", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := gitrepo.Root(context.Background(), nested)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Fatalf("root = %q, want %q", root, dir)
	}
}

func TestArchiveExtractsRefWithoutMutatingWorktree(t *testing.T) {
	dir := initRepo(t)
	writeGitFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeGitFile(t, dir, "main.go", "package main\n")
	runTestGit(t, dir, "add", ".")
	commitTestGit(t, dir, "feat: initial")
	runTestGit(t, dir, "tag", "v1")
	writeGitFile(t, dir, "main.go", "package main\n\nfunc New() {}\n")

	out, cleanup, err := gitrepo.Archive(context.Background(), dir, "v1")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	content, err := os.ReadFile(filepath.Join(out, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "package main\n" {
		t.Fatalf("archived main.go = %q", string(content))
	}
	current, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "package main\n\nfunc New() {}\n" {
		t.Fatalf("worktree main.go = %q", string(current))
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runTestGit(t, dir, "init", "-q")
	runTestGit(t, dir, "config", "user.name", "Test User")
	runTestGit(t, dir, "config", "user.email", "test@example.com")
	return dir
}

func writeGitFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runTestGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func commitTestGit(t *testing.T, root string, message string) {
	t.Helper()
	runTestGit(t, root, "add", ".")
	runTestGit(t, root, "commit", "-q", "-m", message)
}
