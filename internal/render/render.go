package render

import (
	"encoding/json"
	"errors"
	"fmt"
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

func table(w io.Writer, result gocensus.Result) error {
	_, err := fmt.Fprintf(w, "Go Census: %s\n\nFiles\n  Total %d\n",
		displayName(result), result.Files.Total)
	return err
}

func markdown(w io.Writer, result gocensus.Result) error {
	_, err := fmt.Fprintf(w, "# Go Census: %s\n\n## Files\n\n- Total: %d\n",
		displayName(result), result.Files.Total)
	return err
}

func displayName(result gocensus.Result) string {
	if result.ModulePath != "" {
		return result.ModulePath
	}
	return result.Root
}
