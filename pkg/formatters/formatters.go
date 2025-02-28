package formatters

import (
	"fmt"
	"io"
)

const (
	OpCreate = "create"
	OpGet    = "read"
	OpUpdate = "update"
	OpDelete = "delete"
)

// WriteJsonOp generates a response in JSON format
func WriteJsonOp(w io.Writer, object, resource, operation string) (int, error) {
	return fmt.Fprintf(w, "{\"operation\": \"%s\", \"%s\": %s}", operation, resource, object)
}
