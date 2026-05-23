package cli_test

import (
	"bytes"
	"context"
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
