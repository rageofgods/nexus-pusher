package helper

import (
	"fmt"
)

// ContextError defines new struct for error handling to persist calling context
type ContextError struct {
	Context string
	Err     error
}

// Error declares custom error for ContextError struct
func (c *ContextError) Error() string {
	return fmt.Sprintf("%s: %v", c.Context, c.Err)
}
