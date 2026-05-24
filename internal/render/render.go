package render

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/arloliu/gocensus"
	"github.com/arloliu/gocensus/internal/color"
)

type Options struct {
	Style color.Style
}

// Result renders a census result in the requested format.
func Result(w io.Writer, result gocensus.Result, format string) error {
	return ResultWithOptions(w, result, format, Options{Style: color.Plain()})
}

func ResultWithOptions(w io.Writer, result gocensus.Result, format string, opts Options) error {
	switch format {
	case "table":
		return table(w, result, opts.Style)
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "markdown":
		return markdown(w, result)
	default:
		return errors.New("unknown format")
	}
}
