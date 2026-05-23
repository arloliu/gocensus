package render

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/arloliu/gocensus"
)

// Result renders a census result in the requested format.
func Result(w io.Writer, result gocensus.Result, format string) error {
	switch format {
	case "table":
		return table(w, result)
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
