package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	appAuth "github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models/dto" // Keep dto import for ErrorDetail etc.
	appRepos "github.com/yigit/unisphere/internal/app/repositories"
	appServices "github.com/yigit/unisphere/internal/app/services"
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

// HandleAPIError maps application errors to standard HTTP error responses.
func HandleAPIError(ctx *gin.Context, err error) {
	var statusCode int
	var errDetail *dto.ErrorDetail // Use dto.ErrorDetail

	// Default error details
	statusCode = http.StatusInternalServerError
	errDetail = dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected internal error occurred.") // Use dto.NewErrorDetail
	if err != nil {
		errDetail = errDetail.WithDetails(err.Error())
	}

	// --- Specific Service/Repo/Auth Error Mapping ---
	switch {
	// Not Found Errors
	case errors.Is(err, appServices.ErrClassNotFound), errors.Is(err, appRepos.ErrNotFound),
		errors.Is(err, appServices.ErrUserNotFound), errors.Is(err, appServices.ErrPastExamNotFound),
		errors.Is(err, appAuth.ErrResourceNotFound), errors.Is(err, appServices.ErrFacultyNotFound),
		errors.Is(err, appServices.ErrDepartmentNotFound), errors.Is(err, appServices.ErrInstructorNotFound):
		statusCode = http.StatusNotFound
		errDetail = dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Resource Not Found")
		errDetail = errDetail.WithDetails(err.Error())

	// Conflict Errors (Already Exists)
	case errors.Is(err, appServices.ErrEmailAlreadyExists), errors.Is(err, appServices.ErrStudentIDAlreadyExists),
		errors.Is(err, appRepos.ErrFacultyAlreadyExists), errors.Is(err, appRepos.ErrDepartmentAlreadyExists):
		statusCode = http.StatusConflict
		errDetail = dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Resource Already Exists")
		errDetail = errDetail.WithDetails(err.Error())

	// Authentication/Authorization Errors
	case errors.Is(err, appServices.ErrInvalidCredentials), errors.Is(err, appServices.ErrTokenInvalid),
		errors.Is(err, appServices.ErrTokenExpired), errors.Is(err, appServices.ErrTokenRevoked),
		errors.Is(err, appServices.ErrTokenNotFound):
		statusCode = http.StatusUnauthorized
		errDetail = dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication Failed")
		errDetail = errDetail.WithDetails(err.Error())
	case errors.Is(err, appAuth.ErrPermissionDenied), errors.Is(err, appServices.ErrInstructorOnly):
		statusCode = http.StatusForbidden
		errDetail = dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Permission Denied")
		errDetail = errDetail.WithDetails(err.Error())

	// Validation Errors (from binding or service)
	case errors.As(err, &validator.ValidationErrors{}):
		statusCode = http.StatusBadRequest
		errDetail = dto.HandleValidationError(err) // Use dto.HandleValidationError
	case errors.Is(err, appServices.ErrInvalidEmail), errors.Is(err, appServices.ErrInvalidPassword),
		errors.Is(err, appServices.ErrInvalidStudentID), errors.Is(err, appServices.ErrNoteDepartmentNotFound):
		statusCode = http.StatusBadRequest
		errDetail = dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation Failed")
		errDetail = errDetail.WithDetails(err.Error())

	// --- Add other specific error mappings as needed ---

	default:
		// If the error is not specifically handled, log it as an internal error
		if err != nil { // Only log if there is an actual error
			logger.Error().Err(err).Str("path", ctx.Request.URL.Path).Msg("Unhandled internal error")
		}
		// Keep generic message for the client for unexpected errors
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected internal error occurred.") // Use dto.NewErrorDetail
	}

	// Send the standardized error response
	ctx.AbortWithStatusJSON(statusCode, dto.NewErrorResponse(errDetail)) // Use dto.NewErrorResponse
}
