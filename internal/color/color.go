package color

import (
	"strings"
)

type Level int

const (
	LevelPlain Level = iota
	LevelANSI
	LevelRGB
)

type Request struct {
	Mode    string
	NoColor bool
	Environ []string
}

type Style struct {
	level Level
}

func Plain() Style {
	return Style{level: LevelPlain}
}

func ANSI() Style {
	return Style{level: LevelANSI}
}

func RGB() Style {
	return Style{level: LevelRGB}
}

func Resolve(req Request) Style {
	mode := req.Mode
	if mode == "" {
		mode = "auto"
	}
	env := environMap(req.Environ)

	if req.NoColor || mode == "never" {
		return Plain()
	}
	if mode == "always" {
		return RGB()
	}
	if _, ok := env["NO_COLOR"]; ok {
		return Plain()
	}
	if env["CLICOLOR"] == "0" {
		return Plain()
	}
	if force := env["CLICOLOR_FORCE"]; force != "" && force != "0" {
		return RGB()
	}
	if supportsRGB(env) {
		return RGB()
	}
	if supportsANSI(env) {
		return ANSI()
	}
	return Plain()
}

func (s Style) Level() Level {
	return s.level
}

func (s Style) Title(text string) string {
	return s.wrap(text, ansi("1", "36"), rgb(125, 211, 252, true))
}

func (s Style) Section(text string) string {
	return s.wrap(text, ansi("1", "34"), rgb(147, 197, 253, true))
}

func (s Style) Header(text string) string {
	return s.wrap(text, ansi("1", "36"), rgb(186, 230, 253, true))
}

func (s Style) Label(text string) string {
	return s.wrap(text, ansi("36"), rgb(165, 180, 252, false))
}

func (s Style) Metric(text string) string {
	return s.wrap(text, ansi("32"), rgb(203, 213, 225, false))
}

func (s Style) Warn(text string) string {
	return s.wrap(text, ansi("33"), rgb(253, 224, 71, false))
}

func (s Style) Bad(text string) string {
	return s.wrap(text, ansi("31"), rgb(252, 165, 165, false))
}

func (s Style) Muted(text string) string {
	return s.wrap(text, ansi("2"), rgb(148, 163, 184, false))
}

func (s Style) wrap(text string, ansiCode string, rgbCode string) string {
	switch s.level {
	case LevelRGB:
		return rgbCode + text + "\x1b[0m"
	case LevelANSI:
		return ansiCode + text + "\x1b[0m"
	default:
		return text
	}
}

func ansi(codes ...string) string {
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func rgb(r int, g int, b int, bold bool) string {
	style := ""
	if bold {
		style = "\x1b[1m"
	}
	return style + "\x1b[38;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func supportsRGB(env map[string]string) bool {
	colorterm := strings.ToLower(env["COLORTERM"])
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}
	term := strings.ToLower(env["TERM"])
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") {
		return true
	}
	switch strings.ToLower(env["TERM_PROGRAM"]) {
	case "iterm.app", "wezterm", "kitty", "hyper", "apple_terminal", "vscode", "ghostty", "windows_terminal":
		return true
	default:
		return false
	}
}

func supportsANSI(env map[string]string) bool {
	term := env["TERM"]
	if term == "" || term == "dumb" {
		return false
	}
	return true
}

func environMap(environ []string) map[string]string {
	env := make(map[string]string, len(environ))
	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			env[entry] = ""
			continue
		}
		env[name] = value
	}
	return env
}

func StripSGR(text string) string {
	var out strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) {
				if text[i] >= '@' && text[i] <= '~' {
					break
				}
				i++
			}
			continue
		}
		out.WriteByte(text[i])
	}
	return out.String()
}
