package apperrors

import "errors"

// Common errors
var (
	// Resource errors
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrConflict              = errors.New("conflict")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrAccountDisabled    = errors.New("account is disabled")

	// Authorization errors
	ErrPermissionDenied = errors.New("permission denied")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrBadRequest       = errors.New("bad request")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrIdentifierExists   = errors.New("identifier already exists")

	// Class note errors
	ErrClassNoteNotFound = errors.New("class note not found")

	// Past exam errors
	ErrPastExamNotFound = errors.New("past exam not found")
)

// Student Errors
var (
	ErrStudentNotFound        = errors.New("student not found")
	ErrStudentIDAlreadyExists = errors.New("student ID already exists")
	ErrInvalidStudentID       = errors.New("invalid student ID format")
)

// Department Errors
var (
	ErrDepartmentNotFound      = errors.New("department not found")
	ErrDepartmentAlreadyExists = errors.New("department with this name or code already exists")
	ErrDepartmentHasRelations  = errors.New("department has associated data and cannot be deleted")
)

// Faculty Errors
var (
	ErrFacultyNotFound      = errors.New("faculty not found")
	ErrFacultyAlreadyExists = errors.New("faculty with this name or abbreviation already exists")
	ErrFacultyHasRelations  = errors.New("faculty has associated departments and cannot be deleted")
)

// Content Errors
var (
	ErrInvalidFormat = errors.New("invalid token format")
)

// Email verification errors
var (
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrInvalidEmailToken    = errors.New("invalid or expired email verification token")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

// Password reset errors
var (
	ErrInvalidPasswordResetToken = errors.New("invalid or expired password reset token")
	ErrPasswordResetTokenUsed    = errors.New("password reset token has already been used")
)

// NewResourceNotFoundError creates a new custom error for resource not found with a message
func NewResourceNotFoundError(message string) error {
	return &CustomError{
		Err:     ErrResourceNotFound,
		Message: message,
	}
}

// NewConflictError creates a new custom error for conflict situations with a message
func NewConflictError(message string) error {
	return &CustomError{
		Err:     ErrConflict,
		Message: message,
	}
}

// NewForbiddenError creates a new custom error for permission denied with a message
func NewForbiddenError(message string) error {
	return &CustomError{
		Err:     ErrPermissionDenied,
		Message: message,
	}
}

// NewBadRequestError creates a new custom error for bad request with a message
func NewBadRequestError(message string) error {
	return &CustomError{
		Err:     ErrBadRequest,
		Message: message,
	}
}

// Is returns whether target matches any of the errors in errList
// Bu yardımcı fonksiyon, errors.Is() fonksiyonunun birden fazla hatayla kullanımını kolaylaştırır
func Is(err, target error, errList ...error) bool {
	if errors.Is(err, target) {
		return true
	}

	for _, e := range errList {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}

// CustomError represents application-specific errors with additional context
type CustomError struct {
	Err       error
	Message   string
	StatusMsg string
	Code      string
	Details   map[string]interface{}
}

// Error implements error interface
func (e *CustomError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap implements errors.Unwrap interface
func (e *CustomError) Unwrap() error {
	return e.Err
}

// NewCustomError creates a CustomError with underlying error
func NewCustomError(err error, message string) *CustomError {
	return &CustomError{
		Err:     err,
		Message: message,
	}
}

// WithDetails adds context details to the error
func (e *CustomError) WithDetails(details map[string]interface{}) *CustomError {
	e.Details = details
	return e
}

// WithCode adds an error code
func (e *CustomError) WithCode(code string) *CustomError {
	e.Code = code
	return e
}

// WithStatusMsg adds a user-friendly status message
func (e *CustomError) WithStatusMsg(msg string) *CustomError {
	e.StatusMsg = msg
	return e
}
