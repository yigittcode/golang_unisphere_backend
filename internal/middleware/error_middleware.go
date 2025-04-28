package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto" // Keep dto import for ErrorDetail etc.
	"github.com/yigit/unisphere/internal/pkg/apperrors"
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
		c.JSON(404, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Resource not found"),
		})
		return
	case errors.Is(err, apperrors.ErrPermissionDenied):
		c.JSON(403, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeForbidden, "Permission denied"),
		})
		return
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		c.JSON(401, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInvalidCredentials, "Invalid credentials"),
		})
		return
	case errors.Is(err, apperrors.ErrTokenExpired):
		c.JSON(401, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeExpiredToken, "Token expired"),
		})
		return
	case errors.Is(err, apperrors.ErrTokenInvalid):
		c.JSON(401, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInvalidToken, "Invalid token"),
		})
		return
	case errors.Is(err, apperrors.ErrTokenNotFound):
		c.JSON(401, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeTokenNotFound, "Token not found"),
		})
		return
	case errors.Is(err, apperrors.ErrTokenRevoked):
		c.JSON(401, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInvalidToken, "Token revoked"),
		})
		return
	case errors.Is(err, apperrors.ErrValidationFailed):
		c.JSON(400, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed"),
		})
		return
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		c.JSON(409, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Email already exists"),
		})
		return
	case errors.Is(err, apperrors.ErrIdentifierExists):
		c.JSON(409, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Student ID already exists"),
		})
		return
	default:
		// Handle unknown errors
		c.JSON(500, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Internal server error"),
		})
		return
	}
}
