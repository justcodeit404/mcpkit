package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter outputs results as indented JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a JSONFormatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format writes JSON-encoded result to w.
func (f *JSONFormatter) Format(w io.Writer, result any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(result)
}
