package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

/* // REMOVED ErrorCode, ErrorSeverity, ErrorDetail, ErrorResponse structs and helpers
// ErrorCode represents standardized error codes
type ErrorCode string
// ... constants ...
type ErrorSeverity string
// ... constants ...
type ErrorDetail struct { ... }
type ErrorResponse struct { ... }
func NewErrorDetail(code ErrorCode, message string) *ErrorDetail { ... }
func (e *ErrorDetail) WithField(field string) *ErrorDetail { ... }
func (e *ErrorDetail) WithSeverity(severity ErrorSeverity) *ErrorDetail { ... }
func (e *ErrorDetail) WithDetails(details interface{}) *ErrorDetail { ... }
func (e *ErrorDetail) WithDebugInfo(format string, args ...interface{}) *ErrorDetail { ... }
func NewErrorResponse(errorDetail *ErrorDetail) *ErrorResponse { ... }
*/

// --- Central Error Handling Middleware/Function ---

// HandleAPIError handles common API errors and returns appropriate responses
func HandleAPIError(c *gin.Context, err error) {
	// Log the error for debugging
	logger.Debug().Err(err).Str("path", c.Request.URL.Path).Msg("Handling API error")
	
	// Check for specific error types
	switch {
	// Resource Not Found errors
	case errors.Is(err, apperrors.ErrResourceNotFound):
		// Check if it's a custom error with a message
		customErr, ok := err.(*apperrors.CustomError)
		if ok && customErr.Message != "" {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, customErr.Message)))
		} else {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Resource not found")))
		}
		return
	case errors.Is(err, apperrors.ErrClassNoteNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Class note not found")))
		return
	case errors.Is(err, apperrors.ErrPastExamNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Past exam not found")))
		return
	case errors.Is(err, apperrors.ErrUserNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "User not found")))
		return
	case errors.Is(err, apperrors.ErrDepartmentNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Department not found")))
		return
	case errors.Is(err, apperrors.ErrFacultyNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Faculty not found")))
		return
		
	// Authorization/Permission errors
	case errors.Is(err, apperrors.ErrPermissionDenied):
		c.JSON(http.StatusForbidden, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeForbidden, "Permission denied")))
		return
	
	// Authentication errors
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidCredentials, "Invalid credentials")))
		return
	case errors.Is(err, apperrors.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeExpiredToken, "Token expired")))
		return
	case errors.Is(err, apperrors.ErrTokenInvalid):
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidToken, "Invalid token")))
		return
	case errors.Is(err, apperrors.ErrTokenNotFound):
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeTokenNotFound, "Token not found")))
		return
	case errors.Is(err, apperrors.ErrTokenRevoked):
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidToken, "Token revoked")))
		return
	case errors.Is(err, apperrors.ErrAccountDisabled):
		c.JSON(http.StatusForbidden, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeForbidden, "Account is disabled")))
		return
	
	// Validation errors
	case errors.Is(err, apperrors.ErrValidationFailed) || errors.Is(err, apperrors.ErrInvalidPassword):
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	case errors.Is(err, apperrors.ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidEmail, "Invalid email format")))
		return
	case errors.Is(err, apperrors.ErrInvalidFormat):
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid format")))
		return
	case errors.Is(err, apperrors.ErrInvalidStudentID):
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidStudentID, "Invalid student ID format")))
		return
	
	// Resource conflict errors
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Email already exists")))
		return
	case errors.Is(err, apperrors.ErrIdentifierExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Student ID already exists")))
		return
	case errors.Is(err, apperrors.ErrStudentIDAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Student ID already exists")))
		return
	case errors.Is(err, apperrors.ErrResourceAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Resource already exists")))
		return
	case errors.Is(err, apperrors.ErrDepartmentAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Department with this name or code already exists")))
		return
	case errors.Is(err, apperrors.ErrFacultyAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Faculty with this name or abbreviation already exists")))
		return
	
	// Dependency errors
	case errors.Is(err, apperrors.ErrDepartmentHasRelations):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceInvalid, "Department has associated data and cannot be deleted")))
		return
	case errors.Is(err, apperrors.ErrFacultyHasRelations):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceInvalid, "Faculty has associated departments and cannot be deleted")))
		return
	
	default:
		// Log unexpected errors
		logger.Error().Err(err).Str("path", c.Request.URL.Path).Msg("Unexpected error occurred")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected error occurred").WithDetails(err.Error())))
		return
	}
}
