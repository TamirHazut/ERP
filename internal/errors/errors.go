package errors

import (
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	CategoryAuth       ErrorCategory = "AUTH"
	CategoryValidation ErrorCategory = "VALIDATION"
	CategoryNotFound   ErrorCategory = "NOT_FOUND"
	CategoryConflict   ErrorCategory = "CONFLICT"
	CategoryBusiness   ErrorCategory = "BUSINESS"
	CategoryInternal   ErrorCategory = "INTERNAL"
)

// AppError represents a structured application error
type AppError struct {
	Code     string         // e.g., "AUTH_INVALID_CREDENTIALS"
	Message  string         // User-friendly message
	Category ErrorCategory  // AUTH, VALIDATION, NOT_FOUND, etc.
	Details  map[string]any // Optional metadata
	Err      error          // Wrapped underlying error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s - %v", e.Category, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Category, e.Code, e.Message)
}

// Unwrap returns the wrapped error for errors.Is/As support
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches by code
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(key string, value any) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError from an error code definition
func New(def ErrorDef) *AppError {
	return &AppError{
		Code:     def.Code,
		Message:  def.Message,
		Category: def.Category,
		Details:  make(map[string]any),
	}
}

// Wrap creates a new AppError wrapping an existing error
func Wrap(def ErrorDef, err error) *AppError {
	return &AppError{
		Code:     def.Code,
		Message:  def.Message,
		Category: def.Category,
		Details:  make(map[string]any),
		Err:      err,
	}
}

// Auth creates an authentication/authorization error
func Auth(def ErrorDef) *AppError {
	if def.Category == "" {
		def.Category = CategoryAuth
	}
	return New(def)
}

// Validation creates a validation error with optional field information
func Validation(def ErrorDef, fields ...string) *AppError {
	if def.Category == "" {
		def.Category = CategoryValidation
	}
	e := New(def)
	if len(fields) > 0 {
		e.Details["fields"] = fields
	}
	return e
}

// NotFound creates a not found error with optional resource information
func NotFound(def ErrorDef, resourceType string, resourceID any) *AppError {
	if def.Category == "" {
		def.Category = CategoryNotFound
	}
	e := New(def)
	if resourceType != "" {
		e.Details["resource_type"] = resourceType
	}
	if resourceID != nil {
		e.Details["resource_id"] = resourceID
	}
	return e
}

// Conflict creates a conflict error
func Conflict(def ErrorDef) *AppError {
	if def.Category == "" {
		def.Category = CategoryConflict
	}
	return New(def)
}

// Business creates a business rule violation error
func Business(def ErrorDef) *AppError {
	if def.Category == "" {
		def.Category = CategoryBusiness
	}
	return New(def)
}

// Internal creates an internal system error
func Internal(def ErrorDef, err error) *AppError {
	if def.Category == "" {
		def.Category = CategoryInternal
	}
	return Wrap(def, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError attempts to convert an error to an AppError
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if e, ok := err.(*AppError); ok {
		return e, true
	}
	return nil, false
}

// IsCategory checks if an error belongs to a specific category
func IsCategory(err error, category ErrorCategory) bool {
	if e, ok := AsAppError(err); ok {
		return e.Category == category
	}
	return false
}
