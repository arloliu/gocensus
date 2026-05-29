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

func TestRunScanColorAlwaysPrintsSGR(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[38;2;") {
		t.Fatalf("stdout = %q, want RGB SGR", stdout.String())
	}
}

func TestRunScanNoColorSuppressesSGR(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "--color", "always", "--no-color"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("stdout = %q, want no SGR", stdout.String())
	}
}

func TestRunScanJSONIgnoresColor(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "--format", "json", "--color", "always"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Fatalf("stdout = %q, want no SGR", stdout.String())
	}
}

func TestRunScanAcceptsShortFormatFlag(t *testing.T) {
	dir := writeModule(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "-f", "json"}, &stdout, &stderr, "dev")

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

func TestRunReportAcceptsShortOutputFlag(t *testing.T) {
	dir := writeModule(t)
	output := filepath.Join(t.TempDir(), "census.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"report", dir, "-o", output}, &stdout, &stderr, "dev")

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

func TestRunReportTableOutputFileIgnoresColor(t *testing.T) {
	dir := writeModule(t)
	output := filepath.Join(t.TempDir(), "census.txt")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"report", dir, "--format", "table", "--color", "always", "--output", output}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "\x1b[") {
		t.Fatalf("report content = %q, want no SGR", string(content))
	}
}

func TestRunScanAcceptsShortExcludeFlag(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "keep.go", "package main\n")
	writeFile(t, dir, "skip.go", "package main\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "-f", "json", "-x", "skip.go"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	var payload struct {
		Files struct {
			Total int `json:"total"`
		} `json:"files"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload.Files.Total != 1 {
		t.Fatalf("total files = %d, want 1", payload.Files.Total)
	}
}

func TestRunScanExcludesTestdataFixtures(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "testdata/fixture.go", "package fixture\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "-f", "json"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	var payload struct {
		Files struct {
			Total      int `json:"total"`
			Production int `json:"production"`
		} `json:"files"`
		FileMetrics []struct {
			Path string `json:"path"`
		} `json:"file_metrics"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload.Files.Total != 1 || payload.Files.Production != 1 {
		t.Fatalf("files = %#v, want only main.go counted", payload.Files)
	}
	for _, file := range payload.FileMetrics {
		if file.Path == "testdata/fixture.go" {
			t.Fatalf("file_metrics contains testdata fixture: %#v", payload.FileMetrics)
		}
	}
}

func TestRunScanCanIncludeTestdataFixtures(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "testdata/fixture.go", "package fixture\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "-f", "json", "--include-testdata"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	var payload struct {
		Scope string `json:"scope"`
		Files struct {
			Total      int `json:"total"`
			Production int `json:"production"`
		} `json:"files"`
		FileMetrics []struct {
			Path string `json:"path"`
		} `json:"file_metrics"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload.Files.Total != 2 || payload.Files.Production != 2 {
		t.Fatalf("files = %#v, want main.go and testdata fixture counted", payload.Files)
	}
	if payload.Scope != "production excludes generated and mock files; testdata directories included" {
		t.Fatalf("scope = %q, want testdata inclusion scope", payload.Scope)
	}
	if !hasFileMetric(payload.FileMetrics, "testdata/fixture.go") {
		t.Fatalf("file_metrics = %#v, want testdata fixture", payload.FileMetrics)
	}
}

func TestRunScanPrintsScopeForIncludedBuckets(t *testing.T) {
	dir := writeModule(t)
	writeFile(t, dir, "service.pb.go", "// Code generated by protoc; DO NOT EDIT.\npackage main\n")
	writeFile(t, dir, "mock_client.go", "package main\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := cli.Run(context.Background(), []string{"scan", dir, "--include-generated", "--include-mocks"}, &stdout, &stderr, "dev")

	if code != 0 {
		t.Fatalf("Run exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Scope: production includes generated and mock files") {
		t.Fatalf("stdout = %q, want included bucket scope", stdout.String())
	}
	for _, want := range []string{
		"Production Scope",
		"Excluded Generated",
		"Excluded Mocks",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func hasFileMetric(files []struct {
	Path string `json:"path"`
}, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
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
