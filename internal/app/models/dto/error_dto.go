package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standardized error response format
// It's an alias of APIResponse to maintain backward compatibility with Swagger documentation
type ErrorResponse = APIResponse

// ErrorCode represents standardized error codes
type ErrorCode string

// Standard error codes for the application
const (
	ErrorCodeInvalidCredentials    ErrorCode = "AUTH_001"
	ErrorCodeInvalidEmail          ErrorCode = "AUTH_002"
	ErrorCodeInvalidPassword       ErrorCode = "AUTH_003"
	ErrorCodeInvalidStudentID      ErrorCode = "AUTH_004"
	ErrorCodeInvalidToken          ErrorCode = "AUTH_005"
	ErrorCodeExpiredToken          ErrorCode = "AUTH_006"
	ErrorCodeTokenNotFound         ErrorCode = "AUTH_007"
	ErrorCodeUnauthorized          ErrorCode = "AUTH_008"
	ErrorCodeResourceNotFound      ErrorCode = "RES_001"
	ErrorCodeResourceAlreadyExists ErrorCode = "RES_002"
	ErrorCodeResourceInvalid       ErrorCode = "RES_003"
	ErrorCodeValidationFailed      ErrorCode = "VAL_001"
	ErrorCodeInternalServer        ErrorCode = "SRV_001"
	ErrorCodeDatabaseError         ErrorCode = "SRV_002"
	ErrorCodeExternalServiceError  ErrorCode = "SRV_003"
	ErrorCodeBadRequest            ErrorCode = "BAD_REQUEST"
	ErrorCodeForbidden             ErrorCode = "FORBIDDEN"
	ErrorCodeInvalidRequest        ErrorCode = "INVALID_REQUEST"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

// Severity levels
const (
	ErrorSeverityInfo     ErrorSeverity = "INFO"
	ErrorSeverityWarning  ErrorSeverity = "WARNING"
	ErrorSeverityError    ErrorSeverity = "ERROR"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Code      ErrorCode     `json:"code" example:"AUTH_001"`
	Message   string        `json:"message" example:"Email format is invalid, school email required"`
	Field     string        `json:"field,omitempty" example:"email"`
	Severity  ErrorSeverity `json:"severity" example:"ERROR"`
	Details   interface{}   `json:"details,omitempty"`
	DebugInfo string        `json:"debugInfo,omitempty"`
}

// --- Error Helper Functions ---

// NewErrorDetail creates a new error detail
func NewErrorDetail(code ErrorCode, message string) *ErrorDetail {
	return &ErrorDetail{
		Code:     code,
		Message:  message,
		Severity: ErrorSeverityError, // Default to Error severity
	}
}

// WithField adds a field name to the error detail
func (e *ErrorDetail) WithField(field string) *ErrorDetail {
	e.Field = field
	return e
}

// WithSeverity sets the severity level of the error
func (e *ErrorDetail) WithSeverity(severity ErrorSeverity) *ErrorDetail {
	e.Severity = severity
	return e
}

// WithDetails adds additional details to the error
func (e *ErrorDetail) WithDetails(details interface{}) *ErrorDetail {
	e.Details = details
	return e
}

// WithDebugInfo adds debug information (for development/testing only)
func (e *ErrorDetail) WithDebugInfo(format string, args ...interface{}) *ErrorDetail {
	e.DebugInfo = fmt.Sprintf(format, args...)
	return e
}

// NewErrorResponse creates a standard error response
// Returns APIResponse to ensure consistency with success responses
func NewErrorResponse(errorDetail *ErrorDetail) APIResponse {
	return APIResponse{
		Error:     errorDetail,
		Timestamp: time.Now(),
	}
}

// --- Validation Error Handling Helpers ---

// formatValidationError creates a human-readable validation error message
func formatValidationError(e validator.FieldError) string { // Keep internal (lowercase)
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	case "email":
		return e.Field() + " must be a valid email address"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "len":
		return e.Field() + " must have a length of " + e.Param()
	case "numeric":
		return e.Field() + " must contain only numeric characters"
	case "alphanum":
		return e.Field() + " must contain only alphanumeric characters"
	case "uppercase":
		return e.Field() + " must be in uppercase"
	case "url":
		return e.Field() + " must be a valid URL"
	case "gte":
		return e.Field() + " must be greater than or equal to " + e.Param()
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	default:
		return e.Field() + " validation failed: " + e.Tag()
	}
}

// HandleValidationError attempts to parse validation errors (e.g., from Gin binding)
// and convert them into a structured ErrorDetail.
func HandleValidationError(err error) *ErrorDetail {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		// Handle multiple validation errors
		if len(validationErrors) > 0 {
			// For simplicity, return details of the first error
			firstErr := validationErrors[0]
			fieldName := firstErr.Field()
			message := formatValidationError(firstErr) // Use internal helper
			return NewErrorDetail(ErrorCodeValidationFailed, message).WithField(fieldName)
		}
	}
	// Fallback for non-validator errors or if parsing fails
	return NewErrorDetail(ErrorCodeValidationFailed, "Input validation failed").WithDetails(err.Error())
}

// fieldNameFromJSONTag extracts the field name from struct JSON tag
func fieldNameFromJSONTag(fieldName string) string {
	// Try to get the JSON tag name through reflection when possible
	// This is a simplified version - in production you might cache this or use a more robust approach
	return strings.ToLower(fieldName)
}

// validationErrorToMessage converts validator errors to human-readable messages
func validationErrorToMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Should be at least %s characters long", fieldError.Param())
	case "max":
		return fmt.Sprintf("Should be at most %s characters long", fieldError.Param())
	case "len":
		return fmt.Sprintf("Should be exactly %s characters long", fieldError.Param())
	case "numeric":
		return "Should contain only numbers"
	default:
		return fmt.Sprintf("Failed %s validation", fieldError.Tag())
	}
}
