package color_test

import (
	"strings"
	"testing"

	"github.com/arloliu/gocensus/internal/color"
)

func TestResolveHonorsNoColor(t *testing.T) {
	style := color.Resolve(color.Request{NoColor: true, Environ: []string{"CLICOLOR_FORCE=1"}})
	if style.Level() != color.LevelPlain {
		t.Fatalf("level = %v, want plain", style.Level())
	}
}

func TestResolveHonorsNever(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "never", Environ: []string{"CLICOLOR_FORCE=1"}})
	if style.Level() != color.LevelPlain {
		t.Fatalf("level = %v, want plain", style.Level())
	}
}

func TestResolveHonorsAlways(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "always"})
	if style.Level() != color.LevelRGB {
		t.Fatalf("level = %v, want rgb", style.Level())
	}
}

func TestResolveHonorsNoColorEnvironment(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"NO_COLOR=1", "COLORTERM=truecolor"}})
	if style.Level() != color.LevelPlain {
		t.Fatalf("level = %v, want plain", style.Level())
	}
}

func TestResolveHonorsCLICOLORZero(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"CLICOLOR=0", "COLORTERM=truecolor"}})
	if style.Level() != color.LevelPlain {
		t.Fatalf("level = %v, want plain", style.Level())
	}
}

func TestResolveHonorsCLICOLORForce(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"CLICOLOR_FORCE=1"}})
	if style.Level() != color.LevelRGB {
		t.Fatalf("level = %v, want rgb", style.Level())
	}
}

func TestResolveDetectsRGB(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"COLORTERM=truecolor"}})
	if style.Level() != color.LevelRGB {
		t.Fatalf("level = %v, want rgb", style.Level())
	}
	if got := style.Title("Go Census"); !strings.Contains(got, "\x1b[38;2;") {
		t.Fatalf("Title() = %q, want RGB SGR", got)
	}
}

func TestResolveFallsBackToANSI(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"TERM=xterm-256color"}})
	if style.Level() != color.LevelANSI {
		t.Fatalf("level = %v, want ansi", style.Level())
	}
	if got := style.Title("Go Census"); strings.Contains(got, "38;2;") || !strings.Contains(got, "\x1b[") {
		t.Fatalf("Title() = %q, want non-RGB ANSI SGR", got)
	}
}

func TestResolvePlainWithoutTerminalSignals(t *testing.T) {
	style := color.Resolve(color.Request{Mode: "auto", Environ: []string{"TERM=dumb"}})
	if style.Level() != color.LevelPlain {
		t.Fatalf("level = %v, want plain", style.Level())
	}
}

func TestStripSGRRemovesColorSequences(t *testing.T) {
	got := color.StripSGR("\x1b[1;36mGo Census\x1b[0m")
	if got != "Go Census" {
		t.Fatalf("StripSGR() = %q, want Go Census", got)
	}
}

func TestPlainStyleReturnsUnchangedText(t *testing.T) {
	if got := color.Plain().Title("Go Census"); got != "Go Census" {
		t.Fatalf("Title() = %q, want Go Census", got)
	}
}
