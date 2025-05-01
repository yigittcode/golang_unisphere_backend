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
	// Check for specific error types
	switch {
	case errors.Is(err, apperrors.ErrResourceNotFound):
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Resource not found")))
		return
	case errors.Is(err, apperrors.ErrPermissionDenied):
		c.JSON(http.StatusForbidden, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeForbidden, "Permission denied")))
		return
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
	case errors.Is(err, apperrors.ErrValidationFailed) || errors.Is(err, apperrors.ErrInvalidPassword):
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Email already exists")))
		return
	case errors.Is(err, apperrors.ErrIdentifierExists):
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Student ID already exists")))
		return
	default:
		// Log unexpected errors
		logger.Error().Err(err).Msg("Unexpected error occurred")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected error occurred")))
		return
	}
}
