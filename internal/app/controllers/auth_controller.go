package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto" // Ensure DTO import
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/auth"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// AuthController handles authentication related operations
type AuthController struct {
	authService services.AuthService
	jwtService  *auth.JWTService
	logger      *logger.Logger
}

// NewAuthController creates a new AuthController
func NewAuthController(authService services.AuthService, jwtService *auth.JWTService, logger *logger.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// This controller now uses the centralized error handling middleware

// RegisterStudent handles student registration
// @Summary Register a new student
// @Description Create a new student account and return authentication tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterStudentRequest true "Student registration information"
// @Success 201 {object} dto.TokenResponse "Registration successful, tokens returned"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} dto.ErrorResponse "Email or student ID already in use"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register-student [post]
func (c *AuthController) RegisterStudent(ctx *gin.Context) {
	var req dto.RegisterStudentRequest

	// Bind JSON and validate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid registration request payload")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log registration request data for debugging
	c.logger.Info().
		Str("email", req.Email).
		Str("firstName", req.FirstName).
		Str("lastName", req.LastName).
		Str("studentId", req.StudentID).
		Int64("departmentId", req.DepartmentID).
		Msg("Student registration request received")

	if req.GraduationYear != nil {
		c.logger.Info().Int("graduationYear", *req.GraduationYear).Msg("Student registration includes graduation year")
	}

	// Register student
	token, err := c.authService.RegisterStudent(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to register student")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful registration
	c.logger.Info().
		Str("email", req.Email).
		Msg("Student registered successfully")

	// Return success response
	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(token, "Student registration successful"))
}

// RegisterInstructor handles instructor registration
// @Summary Register a new instructor
// @Description Create a new instructor account and return authentication tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterInstructorRequest true "Instructor registration information"
// @Success 201 {object} dto.TokenResponse "Registration successful, tokens returned"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} dto.ErrorResponse "Email already in use"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register-instructor [post]
func (c *AuthController) RegisterInstructor(ctx *gin.Context) {
	var req dto.RegisterInstructorRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid registration request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Log registration request data
	c.logger.Info().
		Str("email", req.Email).
		Str("firstName", req.FirstName).
		Str("lastName", req.LastName).
		Str("title", req.Title).
		Int64("departmentId", req.DepartmentID).
		Msg("Instructor registration request received")

	// Register instructor
	tokenResponse, err := c.authService.RegisterInstructor(ctx, &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to register instructor")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful registration
	c.logger.Info().
		Str("email", req.Email).
		Msg("Instructor registered successfully")

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(tokenResponse, "Instructor registration successful"))
}

// Login handles user login
// @Summary User login
// @Description Authenticate with email and password and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.TokenResponse "Login successful, tokens returned"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid login request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Log login attempt (only email for security)
	c.logger.Info().
		Str("email", req.Email).
		Msg("Login attempt received")

	// Process login request
	tokenResponse, err := c.authService.Login(ctx, &req)
	if err != nil {
		c.logger.Warn().
			Err(err).
			Str("email", req.Email).
			Msg("Login failed")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful login
	c.logger.Info().
		Str("email", req.Email).
		Msg("Login successful")

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(tokenResponse, "Login successful"))
}

// RefreshToken handles token refresh
// @Summary Refresh authentication token
// @Description Generate new access and refresh tokens using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token information"
// @Success 200 {object} dto.TokenResponse "Token refresh successful"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	tokenResponse, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(tokenResponse, "Token refreshed successfully"))
}

// GetProfile handles user profile retrieval
// @Summary Get user profile
// @Description Get profile information for authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BaseUserProfile "User profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile [get]
func (c *AuthController) GetProfile(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		c.logger.Warn().Msg("User ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		c.logger.Error().
			Interface("userID", userID).
			Msg("Failed to convert user ID to int64")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.logger.Info().Int64("userID", userIDInt64).Msg("Fetching profile for user")

	// Get user profile
	profile, err := c.authService.GetProfile(ctx.Request.Context(), userIDInt64)
	if err != nil {
		c.logger.Error().Err(err).Int64("userID", userIDInt64).Msg("Failed to get user profile")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Log profile data for debugging
	c.logger.Info().
		Int64("userID", userIDInt64).
		Interface("faculty", profile.Faculty).
		Msg("Profile data retrieved successfully")

	ctx.JSON(http.StatusOK, profile)
}

// UpdateProfilePhoto handles profile photo upload
// @Summary Update profile photo
// @Description Upload or update a user's profile photo
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Profile photo"
// @Security BearerAuth
// @Success 200 {object} dto.BaseUserProfile "Profile photo updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid file format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile/photo [post]
func (c *AuthController) UpdateProfilePhoto(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		middleware.HandleAPIError(ctx, apperrors.ErrPermissionDenied)
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		middleware.HandleAPIError(ctx, apperrors.ErrInternalServer)
		return
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid file upload")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	profile, err := c.authService.UpdateProfilePhoto(ctx, userID, file)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(profile, "Profile photo updated successfully"))
}

// DeleteProfilePhoto handles profile photo deletion
// @Summary Delete profile photo
// @Description Remove a user's profile photo
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BaseUserProfile "Profile photo deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile/photo [delete]
func (c *AuthController) DeleteProfilePhoto(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		middleware.HandleAPIError(ctx, apperrors.ErrPermissionDenied)
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		middleware.HandleAPIError(ctx, apperrors.ErrInternalServer)
		return
	}

	profile, err := c.authService.DeleteProfilePhoto(ctx, userID)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(profile, "Profile photo deleted successfully"))
}

// UpdateProfile handles user profile updates
// @Summary Update user profile
// @Description Update profile information for authenticated user (name, email)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserProfileRequest true "Profile update data"
// @Success 200 {object} dto.BaseUserProfile "Profile updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile [put]
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		middleware.HandleAPIError(ctx, apperrors.ErrPermissionDenied)
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		middleware.HandleAPIError(ctx, apperrors.ErrInternalServer)
		return
	}

	var req dto.UpdateUserProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	profile, err := c.authService.UpdateUserProfile(ctx, userID, &req)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(profile, "Profile updated successfully"))
}
