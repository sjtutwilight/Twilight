package model

import (
	"fmt"
)

// ProcessorError represents an error that occurred during processing
type ProcessorError struct {
	Operation string // The operation that failed
	Err       error  // The underlying error
}

// NewProcessorError creates a new ProcessorError
func NewProcessorError(operation string, err error) *ProcessorError {
	return &ProcessorError{
		Operation: operation,
		Err:       err,
	}
}

// Error returns a string representation of the error
func (e *ProcessorError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("processor error during %s: unknown error", e.Operation)
	}
	return fmt.Sprintf("processor error during %s: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error
func (e *ProcessorError) Unwrap() error {
	return e.Err
}
