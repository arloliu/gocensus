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

func TestRGBStyleUsesSoftOceanPalette(t *testing.T) {
	style := color.RGB()
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "title", got: style.Title("Go Census"), want: "\x1b[1m\x1b[38;2;125;211;252mGo Census\x1b[0m"},
		{name: "section", got: style.Section("Overview"), want: "\x1b[1m\x1b[38;2;147;197;253mOverview\x1b[0m"},
		{name: "header", got: style.Header("Metric"), want: "\x1b[1m\x1b[38;2;186;230;253mMetric\x1b[0m"},
		{name: "label", got: style.Label("Scope"), want: "\x1b[38;2;165;180;252mScope\x1b[0m"},
		{name: "metric", got: style.Metric("42"), want: "\x1b[38;2;203;213;225m42\x1b[0m"},
		{name: "warn", got: style.Warn("Excluded Generated"), want: "\x1b[38;2;253;224;71mExcluded Generated\x1b[0m"},
		{name: "bad", got: style.Bad("-12"), want: "\x1b[38;2;252;165;165m-12\x1b[0m"},
		{name: "muted", got: style.Muted("note"), want: "\x1b[38;2;148;163;184mnote\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
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
