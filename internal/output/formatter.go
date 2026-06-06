package output

import "io"

// Formatter renders a result to a writer.
type Formatter interface {
	Format(w io.Writer, result any) error
}
