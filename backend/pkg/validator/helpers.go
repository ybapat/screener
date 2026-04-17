package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Common validation patterns
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	slugRegex     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// IsEmpty returns true if there are no validation errors
func (v ValidationErrors) IsEmpty() bool {
	return len(v) == 0
}

// String validation helpers

// ValidateRequired checks if a string is not empty
func ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: "is required",
		}
	}
	return nil
}

// ValidateMinLength checks if a string meets minimum length
func ValidateMinLength(field, value string, min int) *ValidationError {
	if len(value) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", min),
			Value:   len(value),
		}
	}
	return nil
}

// ValidateMaxLength checks if a string does not exceed maximum length
func ValidateMaxLength(field, value string, max int) *ValidationError {
	if len(value) > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d characters", max),
			Value:   len(value),
		}
	}
	return nil
}

// ValidateEmail checks if a string is a valid email address
func ValidateEmail(field, value string) *ValidationError {
	if !emailRegex.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid email address",
		}
	}
	return nil
}

// ValidateAlphanumeric checks if a string contains only letters and numbers
func ValidateAlphanumeric(field, value string) *ValidationError {
	if !alphanumRegex.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must contain only letters and numbers",
		}
	}
	return nil
}

// ValidateSlug checks if a string is a valid URL slug
func ValidateSlug(field, value string) *ValidationError {
	if !slugRegex.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid slug (lowercase letters, numbers, and hyphens only)",
		}
	}
	return nil
}

// ValidateOneOf checks if a value is in a list of allowed values
func ValidateOneOf(field, value string, allowed []string) *ValidationError {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
		Value:   value,
	}
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(field, value string) *ValidationError {
	if len(value) < 8 {
		return &ValidationError{
			Field:   field,
			Message: "must be at least 8 characters long",
		}
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return &ValidationError{
			Field:   field,
			Message: "must contain at least one uppercase letter, one lowercase letter, and one digit",
		}
	}

	return nil
}

// Numeric validation helpers

// ValidateRange checks if an integer is within a range
func ValidateRange(field string, value, min, max int) *ValidationError {
	if value < min || value > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d", min, max),
			Value:   value,
		}
	}
	return nil
}

// ValidatePositive checks if an integer is positive
func ValidatePositive(field string, value int) *ValidationError {
	if value <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be positive",
			Value:   value,
		}
	}
	return nil
}

// ValidateNonNegative checks if an integer is non-negative
func ValidateNonNegative(field string, value int) *ValidationError {
	if value < 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be non-negative",
			Value:   value,
		}
	}
	return nil
}

// Time validation helpers

// ValidateNotFuture checks if a time is not in the future
func ValidateNotFuture(field string, value time.Time) *ValidationError {
	if value.After(time.Now()) {
		return &ValidationError{
			Field:   field,
			Message: "cannot be in the future",
			Value:   value,
		}
	}
	return nil
}

// ValidateAfter checks if a time is after another time
func ValidateAfter(field string, value, after time.Time) *ValidationError {
	if !value.After(after) {
		return &ValidationError{
			Field:   field,
			Message: "must be after the start time",
			Value:   value,
		}
	}
	return nil
}

// ValidateDuration checks if a duration is within a range
func ValidateDuration(field string, duration time.Duration, min, max time.Duration) *ValidationError {
	if duration < min || duration > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %s and %s", min, max),
			Value:   duration,
		}
	}
	return nil
}

// Collection validation helpers

// ValidateSliceLength checks if a slice has a valid length
func ValidateSliceLength[T any](field string, value []T, min, max int) *ValidationError {
	length := len(value)
	if length < min || length > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must contain between %d and %d items", min, max),
			Value:   length,
		}
	}
	return nil
}

// ValidateUnique checks if all elements in a slice are unique
func ValidateUnique[T comparable](field string, value []T) *ValidationError {
	seen := make(map[T]bool)
	for _, item := range value {
		if seen[item] {
			return &ValidationError{
				Field:   field,
				Message: "must contain unique values",
			}
		}
		seen[item] = true
	}
	return nil
}

// Utility functions

// CollectErrors collects non-nil validation errors
func CollectErrors(errors ...*ValidationError) ValidationErrors {
	var result ValidationErrors
	for _, err := range errors {
		if err != nil {
			result = append(result, *err)
		}
	}
	return result
}

// AddError adds a validation error to a collection
func AddError(errors ValidationErrors, err *ValidationError) ValidationErrors {
	if err != nil {
		return append(errors, *err)
	}
	return errors
}
