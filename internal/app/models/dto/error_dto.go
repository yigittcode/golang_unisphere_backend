package dto

import (
	"fmt"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

// Standard error codes for the application
const (
	// Authentication errors
	ErrorCodeInvalidCredentials ErrorCode = "AUTH_001"
	ErrorCodeInvalidEmail       ErrorCode = "AUTH_002"
	ErrorCodeInvalidPassword    ErrorCode = "AUTH_003"
	ErrorCodeInvalidStudentID   ErrorCode = "AUTH_004"
	ErrorCodeInvalidToken       ErrorCode = "AUTH_005"
	ErrorCodeExpiredToken       ErrorCode = "AUTH_006"
	ErrorCodeTokenNotFound      ErrorCode = "AUTH_007"
	ErrorCodeUnauthorized       ErrorCode = "AUTH_008"

	// Resource errors
	ErrorCodeResourceNotFound      ErrorCode = "RES_001"
	ErrorCodeResourceAlreadyExists ErrorCode = "RES_002"
	ErrorCodeResourceInvalid       ErrorCode = "RES_003"

	// Validation errors
	ErrorCodeValidationFailed ErrorCode = "VAL_001"

	// Server errors
	ErrorCodeInternalServer       ErrorCode = "SRV_001"
	ErrorCodeDatabaseError        ErrorCode = "SRV_002"
	ErrorCodeExternalServiceError ErrorCode = "SRV_003"
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

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Success   bool         `json:"success" example:"false"`
	Error     *ErrorDetail `json:"error"`
	Timestamp time.Time    `json:"timestamp" example:"2025-04-23T12:01:05.123Z"`
}

// NewErrorDetail creates a new error detail
func NewErrorDetail(code ErrorCode, message string) *ErrorDetail {
	return &ErrorDetail{
		Code:     code,
		Message:  message,
		Severity: ErrorSeverityError,
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
func NewErrorResponse(errorDetail *ErrorDetail) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     errorDetail,
		Timestamp: time.Now(),
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ErrorDetail `json:"errors"`
}

// NewValidationErrors creates a new validation errors container
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ErrorDetail, 0),
	}
}

// AddError adds a validation error to the container
func (v *ValidationErrors) AddError(field, message string) *ValidationErrors {
	v.Errors = append(v.Errors, ErrorDetail{
		Code:     ErrorCodeValidationFailed,
		Message:  message,
		Field:    field,
		Severity: ErrorSeverityError,
	})
	return v
}

// HasErrors checks if there are any validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}
