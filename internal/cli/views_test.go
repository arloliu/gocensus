package cli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/cli"
)

func TestRunTestsRejectsOutputFlag(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"tests", dir, "--output", "tests.txt"}, &stdout, &stderr, "dev")

	if code != 2 {
		t.Fatalf("Run exit code = %d, want 2; stdout=%q", code, stdout.String())
	}
	if !strings.Contains(stderr.String(), "--output") {
		t.Fatalf("stderr = %q, want mention of --output", stderr.String())
	}
}

func TestRunPackagesPrintsPackageView(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"packages", dir}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Packages") {
		t.Fatalf("stdout = %q, want package view", stdout.String())
	}
}

func TestRunPackagesColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"packages", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunPackagesSortsByProductionLines(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "small/small.go", `package small

func One() {}
`)
	writeFile(t, dir, "large/large.go", `package large

func One() {}

func Two() {}

func Three() {}
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"packages", dir, "--sort", "prod-lines"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	largeIndex := strings.Index(stdout.String(), "large")
	smallIndex := strings.Index(stdout.String(), "small")
	if largeIndex < 0 || smallIndex < 0 {
		t.Fatalf("stdout = %q, want large and small packages", stdout.String())
	}
	if largeIndex > smallIndex {
		t.Fatalf("stdout = %q, want large before small", stdout.String())
	}
}

func TestRunPackagesAcceptsShortSortFlag(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "small/small.go", `package small

func One() {}
`)
	writeFile(t, dir, "large/large.go", `package large

func One() {}

func Two() {}

func Three() {}
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"packages", dir, "-s", "prod-lines"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	largeIndex := strings.Index(stdout.String(), "large")
	smallIndex := strings.Index(stdout.String(), "small")
	if largeIndex < 0 || smallIndex < 0 {
		t.Fatalf("stdout = %q, want large and small packages", stdout.String())
	}
	if largeIndex > smallIndex {
		t.Fatalf("stdout = %q, want large before small", stdout.String())
	}
}

func TestRunFilesAcceptsShortTopFlag(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "one.go", "package main\n")
	writeFile(t, dir, "two.go", "package main\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"files", dir, "-n", "1"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("stdout = %q, want header plus one file", stdout.String())
	}
}

func TestRunFilesColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "one.go", "package main\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"files", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunTestsPrintsKnownCaseTotals(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "main_test.go", `package main

import "testing"

func TestOne(t *testing.T) {
	t.Run("child", func(t *testing.T) {})
}

func TestTable(t *testing.T) {
	cases := []struct{ name string }{{"one"}, {"two"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {})
	}
}
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"tests", dir}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Known Test Cases",
		"Top-level Tests",
		"Static Subtests",
		"5",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunTestsColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"tests", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunDiffComparesGitRefs(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	runGit(t, dir, "tag", "base")
	writeFile(t, dir, "main.go", "package main\n\nfunc NewFeature() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author Two", "two@example.com", "2026-01-04T12:00:00Z", "feat: expand")
	runGit(t, dir, "tag", "head")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"diff", dir, "--base", "base", "--head", "head"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Diff:", "Base: base", "Head: head", "Scope:", "Production Effective", "+"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func TestRunDiffColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	runGit(t, dir, "tag", "base")
	writeFile(t, dir, "main.go", "package main\n\nfunc NewFeature() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author Two", "two@example.com", "2026-01-04T12:00:00Z", "feat: expand")
	runGit(t, dir, "tag", "head")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"diff", dir, "--base", "base", "--head", "head", "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunHotspotsRanksProductionFiles(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "small.go", "package main\n")
	writeFile(t, dir, "large.go", "package main\n\nfunc One() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	writeFile(t, dir, "large.go", "package main\n\nfunc One() {}\n\nfunc Two() {}\n\nfunc Three() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author Two", "two@example.com", "2026-01-04T12:00:00Z", "feat: expand")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"hotspots", dir, "-n", "1"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Hotspots:", "Scope:", "Sorted by: score", "Score", "Eff Lines", "Churn", "large.go"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	if strings.Contains(output, "small.go") {
		t.Fatalf("stdout = %q, did not want small.go when top is 1", output)
	}
}

func TestRunHotspotsColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"hotspots", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunHotspotsCanIncludeGeneratedAndMocks(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "service.pb.go", "package main\n\nfunc GeneratedOne() {}\n\nfunc GeneratedTwo() {}\n")
	writeFile(t, dir, "mock_client.go", "package main\n\nfunc MockOne() {}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Author One", "one@example.com", "2026-01-03T12:00:00Z", "feat: initial")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"hotspots", dir, "--include-generated", "--include-mocks", "-n", "0"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Scope: production includes generated and mock files",
		"service.pb.go",
		"mock_client.go",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func TestRunWhoRanksContributorsByRemovedLines(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "app.txt", "one\ntwo\nthree\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "feat: add app")
	writeFile(t, dir, "app.txt", "one\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Bob", "bob@example.com", "2026-01-04T12:00:00Z", "fix: close issue #12")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "--by", "removed"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Who", "Author", "Commits", "Feat", "Fix", "Refactor", "Added", "Removed", "Churn"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	bobIndex := strings.Index(output, "Bob")
	aliceIndex := strings.Index(output, "Alice")
	if bobIndex < 0 || aliceIndex < 0 {
		t.Fatalf("stdout = %q, want Bob and Alice", output)
	}
	if bobIndex > aliceIndex {
		t.Fatalf("stdout = %q, want Bob before Alice", output)
	}
}

func TestRunWhoColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "app.txt", "one\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "feat: add app")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunWhoGoOnlyDefaultsToHumanAuthoredGo(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "feat: add main")
	writeFile(t, dir, "service.pb.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Bob", "bob@example.com", "2026-01-04T12:00:00Z", "feat: add generated")
	writeFile(t, dir, "mock_client.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Carol", "carol@example.com", "2026-01-05T12:00:00Z", "feat: add mock")
	writeFile(t, dir, "README.md", "docs\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Dan", "dan@example.com", "2026-01-06T12:00:00Z", "docs: add docs")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "--go-only", "-n", "0"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Scope: human-authored Go files",
		"generated and mock paths excluded",
		"Feature/Fix/Refactor: commit-message heuristics",
		"Line metrics: git log --numstat",
		"Alice",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	for _, notWant := range []string{"Bob", "Carol", "Dan"} {
		if strings.Contains(output, notWant) {
			t.Fatalf("stdout = %q, did not want %q", output, notWant)
		}
	}
}

func TestRunWhoGoOnlyCanIncludeGeneratedAndMocks(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "main.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "feat: add main")
	writeFile(t, dir, "service.pb.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Bob", "bob@example.com", "2026-01-04T12:00:00Z", "feat: add generated")
	writeFile(t, dir, "mocks/client.go", "package main\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Carol", "carol@example.com", "2026-01-05T12:00:00Z", "feat: add mock")
	writeFile(t, dir, "README.md", "docs\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Dan", "dan@example.com", "2026-01-06T12:00:00Z", "docs: add docs")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "--go-only", "--include-generated", "--include-mocks", "-n", "0"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Scope: all Go files", "Alice", "Bob", "Carol"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	if strings.Contains(output, "Dan") {
		t.Fatalf("stdout = %q, did not want Dan", output)
	}
}

func TestRunWhoExcludesGeneratedAndMocksByDefault(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "README.md", "docs\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "docs: add readme")
	writeFile(t, dir, "docs/generated/report.md", "generated report\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Bob", "bob@example.com", "2026-01-04T12:00:00Z", "docs: add generated report")
	writeFile(t, dir, "fixtures/mock_payload.json", "{}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Carol", "carol@example.com", "2026-01-05T12:00:00Z", "test: add mock payload")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "-n", "0"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Scope: all Git-tracked files, generated and mock paths excluded",
		"Alice",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
	for _, notWant := range []string{"Bob", "Carol"} {
		if strings.Contains(output, notWant) {
			t.Fatalf("stdout = %q, did not want %q", output, notWant)
		}
	}
}

func TestRunWhoCanIncludeGeneratedAndMocksAcrossAllTrackedFiles(t *testing.T) {
	dir := writeGitRepo(t)
	writeFile(t, dir, "README.md", "docs\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Alice", "alice@example.com", "2026-01-03T12:00:00Z", "docs: add readme")
	writeFile(t, dir, "docs/generated/report.md", "generated report\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Bob", "bob@example.com", "2026-01-04T12:00:00Z", "docs: add generated report")
	writeFile(t, dir, "fixtures/mock_payload.json", "{}\n")
	runGit(t, dir, "add", ".")
	commitGit(t, dir, "Carol", "carol@example.com", "2026-01-05T12:00:00Z", "test: add mock payload")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"who", dir, "--include-generated", "--include-mocks", "-n", "0"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"Scope: all Git-tracked files, including generated and mock paths",
		"Alice",
		"Bob",
		"Carol",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func writeGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"HOME="+filepath.ToSlash(t.TempDir()),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func commitGit(t *testing.T, dir string, name string, email string, date string, message string) {
	t.Helper()

	cmd := exec.Command("git", "commit", "--author", name+" <"+email+">", "-m", message)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_COMMITTER_NAME="+name,
		"GIT_COMMITTER_EMAIL="+email,
		"GIT_AUTHOR_DATE="+date,
		"GIT_COMMITTER_DATE="+date,
		"HOME="+filepath.ToSlash(t.TempDir()),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, output)
	}
}
