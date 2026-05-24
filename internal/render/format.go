package render

import (
	"fmt"
	"strings"

	"github.com/arloliu/gocensus"
)

func displayName(result gocensus.Result) string {
	if result.ModulePath != "" {
		return result.ModulePath
	}
	return result.Root
}

func formatInt(value int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	text := fmt.Sprintf("%d", value)
	if len(text) <= 3 {
		return sign + text
	}

	firstGroup := len(text) % 3
	if firstGroup == 0 {
		firstGroup = 3
	}

	var out strings.Builder
	out.WriteString(text[:firstGroup])
	for i := firstGroup; i < len(text); i += 3 {
		out.WriteString("," + text[i:i+3])
	}
	return sign + out.String()
}

func pct(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func ratio(value float64) string {
	return fmt.Sprintf("%.2f:1", value)
}
