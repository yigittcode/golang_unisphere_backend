// Package controllers handles HTTP request handling
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/auth"
)

// AuthController handles authentication related operations
type AuthController struct {
	authService services.AuthService
	userRepo    repositories.IUserRepository
	jwtService  *auth.JWTService
	logger      zerolog.Logger
}

// NewAuthController creates a new AuthController
func NewAuthController(authService services.AuthService, userRepo repositories.IUserRepository, jwtService *auth.JWTService, logger zerolog.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		userRepo:    userRepo,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// RegisterStudent handles student registration
// @Summary Register a new student
// @Description Creates a new student account with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterStudentRequest true "Student registration information"
// @Success 201 {object} dto.APIResponse{data=dto.AuthResponse} "Student registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 409 {object} dto.ErrorResponse "Email or student ID already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register/student [post]
func (c *AuthController) RegisterStudent(ctx *gin.Context) {
	var req dto.RegisterStudentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid student registration request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
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
	ctx.JSON(http.StatusCreated, token)
}

// RegisterInstructor handles instructor registration
// @Summary Register a new instructor
// @Description Creates a new instructor account with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterInstructorRequest true "Instructor registration information"
// @Success 201 {object} dto.APIResponse{data=dto.AuthResponse} "Instructor registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register/instructor [post]
func (c *AuthController) RegisterInstructor(ctx *gin.Context) {
	var req dto.RegisterInstructorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid instructor registration request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Log registration request data for debugging
	c.logger.Info().
		Str("email", req.Email).
		Str("firstName", req.FirstName).
		Str("lastName", req.LastName).
		Int64("departmentId", req.DepartmentID).
		Str("title", req.Title).
		Msg("Instructor registration request received")

	// Register instructor
	token, err := c.authService.RegisterInstructor(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to register instructor")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful registration
	c.logger.Info().
		Str("email", req.Email).
		Msg("Instructor registered successfully")

	// Return success response
	ctx.JSON(http.StatusCreated, token)
}

// Login handles user login
// @Summary Login user
// @Description Authenticates a user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User login credentials"
// @Success 200 {object} dto.APIResponse{data=dto.TokenResponse} "Login successful"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 403 {object} dto.ErrorResponse "Account is locked or disabled"
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

	// Return success response
	ctx.JSON(http.StatusOK, tokenResponse)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token information"
// @Success 200 {object} dto.APIResponse{data=dto.TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid refresh token request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Process refresh token request
	tokenResponse, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Token refresh failed")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful token refresh
	c.logger.Info().Msg("Token refreshed successfully")

	// Return success response
	ctx.JSON(http.StatusOK, tokenResponse)
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Retrieves the profile information of the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /auth/profile [get]
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated"),
		})
		return
	}

	user, err := c.authService.GetUserByID(ctx, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get user profile"),
		})
		return
	}

	// Create response
	response := dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.RoleType),
		DepartmentID: user.DepartmentID,
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: response,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Profile update details"
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /auth/profile [put]
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated"),
		})
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format"),
		})
		return
	}

	// Update profile
	err := c.authService.UpdateProfile(ctx, userID.(int64), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to update profile"),
		})
		return
	}

	// Get updated user
	user, err := c.authService.GetUserByID(ctx, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get updated profile"),
		})
		return
	}

	// Create response
	response := dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.RoleType),
		DepartmentID: user.DepartmentID,
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: response,
	})
}

// UpdateProfilePhoto godoc
// @Summary Update profile photo
// @Description Update the current user's profile photo
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Profile photo file"
// @Success 200 {object} dto.APIResponse{data=dto.FileResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /auth/profile/photo [put]
func (c *AuthController) UpdateProfilePhoto(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated"),
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid or missing file"),
		})
		return
	}

	// Update profile photo
	err = c.authService.UpdateProfilePhoto(ctx, userID.(int64), file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to update profile photo"),
		})
		return
	}

	// Get updated user
	user, err := c.authService.GetUserByID(ctx, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get updated profile"),
		})
		return
	}

	// Create response
	response := dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.RoleType),
		DepartmentID: user.DepartmentID,
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: response,
	})
}

// DeleteProfilePhoto godoc
// @Summary Delete profile photo
// @Description Delete the current user's profile photo
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.UserResponse}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /auth/profile/photo [delete]
func (c *AuthController) DeleteProfilePhoto(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated"),
		})
		return
	}

	// Delete profile photo
	err := c.authService.DeleteProfilePhoto(ctx, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete profile photo"),
		})
		return
	}

	// Get updated user
	user, err := c.authService.GetUserByID(ctx, userID.(int64))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.APIResponse{
			Error: dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get updated profile"),
		})
		return
	}

	// Create response
	response := dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.RoleType),
		DepartmentID: user.DepartmentID,
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: response,
	})
}
