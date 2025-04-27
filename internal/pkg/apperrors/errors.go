package apperrors

import "errors"

// Common Errors
var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("resource not found")
)

// User Errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Student Errors
var (
	ErrStudentNotFound   = errors.New("student not found")
	ErrIdentifierExists  = errors.New("student identifier already in use")
	ErrInvalidIdentifier = errors.New("invalid student identifier format")
)

// Authentication Errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("invalid password format")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenInvalid       = errors.New("invalid token")
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
	ErrPastExamNotFound  = errors.New("past exam not found")
	ErrClassNoteNotFound = errors.New("class note not found")
)

// Permissions and Authorization Errors
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrInstructorOnly   = errors.New("this operation is restricted to instructors")
)

// Validation
var (
	ErrValidationFailed = errors.New("validation failed")
)

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
