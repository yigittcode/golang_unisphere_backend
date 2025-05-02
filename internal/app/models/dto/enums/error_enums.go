package enums

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
	ErrorCodeConflict              ErrorCode = "RES_004"
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