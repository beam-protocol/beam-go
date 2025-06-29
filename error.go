package beam

import "fmt"

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// NewError creates a new ValidationError with the specified field and message
func NewError(field, message string) error {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
