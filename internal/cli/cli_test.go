package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/cli"
)

func TestRunDefaultsToScan(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{dir}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Go Census") {
		t.Fatalf("stdout %q does not contain report title", stdout.String())
	}
	if !strings.Contains(stdout.String(), "example.com/app") {
		t.Fatalf("stdout %q does not contain module path", stdout.String())
	}
}

func TestRunScanJSON(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "--format", "json"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	var payload struct {
		ModulePath string `json:"module_path"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload.ModulePath != "example.com/app" {
		t.Fatalf("module_path = %q, want example.com/app", payload.ModulePath)
	}
}

func TestRunRootHelpIncludesCommandsAndCommonFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"--help"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"scan",
		"tests",
		"--no-gitignore",
		"--include-generated",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunSubcommandHelpIncludesCommandFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", "--help"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"scan",
		"--format",
		"--output",
		"--exclude",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunHelpCommandShowsSubcommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"help", "tests"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"Usage:",
		"tests",
		"Test/benchmark/example inventory",
		"--no-gitignore",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

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

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"version"}, &stdout, &stderr, "v0.0.0-test")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "gocensus v0.0.0-test" {
		t.Fatalf("stdout = %q, want version line", stdout.String())
	}
}

func TestRunReportWritesMarkdownFile(t *testing.T) {
	dir := writeModule(t)
	output := filepath.Join(t.TempDir(), "census.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"report", dir, "--output", output}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "# Go Census: example.com/app") {
		t.Fatalf("report content = %q", string(content))
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

func writeModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")
	return dir
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
