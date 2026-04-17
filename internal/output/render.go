package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func WriteText(w io.Writer, lines ...string) error {
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}

func WriteError(w io.Writer, asJSON bool, err error) error {
	if asJSON {
		return WriteJSON(w, map[string]any{"error": err.Error()})
	}

	_, writeErr := fmt.Fprintf(w, "Error: %s\n", err)
	return writeErr
}
