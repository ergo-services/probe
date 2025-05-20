package unit

import (
	"bytes"
	"encoding/json"
	"strings"
)

func marshalWithoutEscaping(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // Prevent escaping <, >, &, etc.
	enc.SetIndent("", "  ")  // Indent with 2 spaces
	err := enc.Encode(v)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil // Remove trailing newline
}
