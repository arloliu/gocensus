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

func writeModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
