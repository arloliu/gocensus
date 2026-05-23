package main

import (
	"runtime/debug"
	"testing"
)

func TestEffectiveVersionUsesExplicitBuildVersion(t *testing.T) {
	got := effectiveVersion("v0.1.0", &debug.BuildInfo{
		Main: debug.Module{Version: "v0.2.0"},
	})
	if got != "v0.1.0" {
		t.Fatalf("effectiveVersion = %q, want v0.1.0", got)
	}
}

func TestEffectiveVersionUsesModuleVersionForDevBuild(t *testing.T) {
	got := effectiveVersion("dev", &debug.BuildInfo{
		Main: debug.Module{Version: "v0.2.0"},
	})
	if got != "v0.2.0" {
		t.Fatalf("effectiveVersion = %q, want v0.2.0", got)
	}
}

func TestEffectiveVersionFallsBackToDev(t *testing.T) {
	got := effectiveVersion("dev", &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
	})
	if got != "dev" {
		t.Fatalf("effectiveVersion = %q, want dev", got)
	}
}
