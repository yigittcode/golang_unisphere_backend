package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/auth"
)

// AuthController handles authentication related operations
type AuthController struct {
	authService *services.AuthService
	jwtService  *auth.JWTService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *services.AuthService, jwtService *auth.JWTService) *AuthController {
	return &AuthController{
		authService: authService,
		jwtService:  jwtService,
	}
}

// handleError is a helper function to handle common error scenarios and send appropriate responses
func handleError(ctx *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorCode := dto.ErrorCodeInternalServer
	errorMessage := "An unexpected error occurred"
	errorDetails := err.Error()

	// Handle specific errors
	switch {
	// Validation errors
	case errors.Is(err, services.ErrInvalidEmail):
		statusCode = http.StatusBadRequest
		errorCode = dto.ErrorCodeValidationFailed
		errorMessage = "Invalid email format"
		errorDetails = "Please provide a valid email address"
	case errors.Is(err, services.ErrInvalidPassword):
		statusCode = http.StatusBadRequest
		errorCode = dto.ErrorCodeValidationFailed
		errorMessage = "Invalid password format"
		errorDetails = "Password must be at least 8 characters and contain at least one letter and one number"
	case errors.Is(err, services.ErrInvalidStudentID):
		statusCode = http.StatusBadRequest
		errorCode = dto.ErrorCodeValidationFailed
		errorMessage = "Invalid student ID format"
		errorDetails = "Student ID must be exactly 8 digits"

	// Conflict errors
	case errors.Is(err, services.ErrEmailAlreadyExists):
		statusCode = http.StatusConflict
		errorCode = dto.ErrorCodeResourceAlreadyExists
		errorMessage = "Registration failed: email already in use"
		errorDetails = "The provided email address is already registered"
	case errors.Is(err, services.ErrStudentIDAlreadyExists):
		statusCode = http.StatusConflict
		errorCode = dto.ErrorCodeResourceAlreadyExists
		errorMessage = "Registration failed: student ID already in use"
		errorDetails = "The provided student ID is already registered"

	// Authentication errors
	case errors.Is(err, services.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		errorCode = dto.ErrorCodeUnauthorized
		errorMessage = "Authentication failed"
		errorDetails = "Invalid email or password"
	case errors.Is(err, services.ErrTokenNotFound),
		errors.Is(err, services.ErrTokenExpired),
		errors.Is(err, services.ErrTokenRevoked),
		errors.Is(err, services.ErrTokenInvalid):
		statusCode = http.StatusUnauthorized
		errorCode = dto.ErrorCodeUnauthorized
		errorMessage = "Invalid or expired token"
		errorDetails = "Please login again to obtain a new token"

	// Not found errors
	case errors.Is(err, services.ErrUserNotFound):
		statusCode = http.StatusNotFound
		errorCode = dto.ErrorCodeResourceNotFound
		errorMessage = "User not found"
		errorDetails = "The requested user does not exist"
	}

	errorDetail := dto.NewErrorDetail(errorCode, errorMessage)
	errorDetail = errorDetail.WithDetails(errorDetails)
	ctx.JSON(statusCode, dto.NewErrorResponse(errorDetail))
}

// RegisterStudent handles student registration
// @Summary Register a new student
// @Description Create a new student account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterStudentRequest true "Student registration information"
// @Success 201 {object} dto.APIResponse "Student successfully registered"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} dto.ErrorResponse "Email or student ID already in use"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register-student [post]
func (c *AuthController) RegisterStudent(ctx *gin.Context) {
	// Parse request body
	var req dto.RegisterStudentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid registration data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to register student
	tokenResponse, err := c.authService.RegisterStudent(ctx, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Return successful response
	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success:   true,
		Message:   "Student registration successful",
		Data:      tokenResponse,
		Timestamp: time.Now(),
	})
}

// RegisterInstructor handles instructor registration
// @Summary Register a new instructor
// @Description Create a new instructor account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterInstructorRequest true "Instructor registration information"
// @Success 201 {object} dto.APIResponse "Instructor successfully registered"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} dto.ErrorResponse "Email already in use"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register-instructor [post]
func (c *AuthController) RegisterInstructor(ctx *gin.Context) {
	// Parse request body
	var req dto.RegisterInstructorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid registration data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to register instructor
	tokenResponse, err := c.authService.RegisterInstructor(ctx, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Return successful response
	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success:   true,
		Message:   "Instructor registration successful",
		Data:      tokenResponse,
		Timestamp: time.Now(),
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.APIResponse "Login successful"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	// Parse request body
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid login data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to authenticate user
	tokenResponse, err := c.authService.Login(ctx, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Return successful response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Login successful",
		Data:      tokenResponse,
		Timestamp: time.Now(),
	})
}

// RefreshToken handles token refresh
// @Summary Refresh authentication token
// @Description Generate new access and refresh tokens using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token information"
// @Success 200 {object} dto.APIResponse "Token refresh successful"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	// Parse request body
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid token refresh request")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to refresh token
	tokenResponse, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Return successful response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Token refreshed successfully",
		Data:      tokenResponse,
		Timestamp: time.Now(),
	})
}

// GetProfile handles user profile retrieval
// @Summary Get user profile
// @Description Get profile information for authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.UserProfile} "User profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile [get]
func (c *AuthController) GetProfile(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
		errorDetail = errorDetail.WithDetails("User ID not found in request context")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert user ID to int64
	userID, ok := userIDInterface.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid user ID format")
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to get user profile
	profile, err := c.authService.GetProfile(ctx, userID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Return successful response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Profile retrieved successfully",
		Data:      profile,
		Timestamp: time.Now(),
	})
}
