// Package controllers handles HTTP request handling
package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
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

// Register handles user registration
// @Summary Register a new user
// @Description Creates a new user account (student or instructor) with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration information"
// @Success 201 {object} dto.APIResponse{data=dto.TokenResponse} "User registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or invalid role type"
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	c.logger.Debug().Msg("Register endpoint called")
	
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid registration request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Validate role type
	validRoles := []string{string(models.RoleStudent), string(models.RoleInstructor)}
	roleIsValid := false
	
	for _, role := range validRoles {
		if string(req.RoleType) == role {
			roleIsValid = true
			break
		}
	}
	
	if !roleIsValid {
		c.logger.Warn().Str("roleType", string(req.RoleType)).Msg("Invalid role type")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid role type")
		errorDetail = errorDetail.WithDetails("Role type must be either STUDENT or INSTRUCTOR")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Validate password
	if err := c.authService.ValidatePassword(req.Password); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid password format")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid password format")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Validate email format (school domain)
	if !strings.HasSuffix(req.Email, ".edu.tr") {
		c.logger.Warn().Str("email", req.Email).Msg("Invalid email domain")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInvalidEmail, "Invalid email domain")
		errorDetail = errorDetail.WithDetails("Email must be from a .edu.tr domain")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Log registration request data for debugging
	c.logger.Info().
		Str("email", req.Email).
		Str("firstName", req.FirstName).
		Str("lastName", req.LastName).
		Int64("departmentId", req.DepartmentID).
		Str("roleType", string(req.RoleType)).
		Msg("User registration request received")

	// Register user
	tokenResponse, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to register user")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful registration
	c.logger.Info().
		Str("email", req.Email).
		Msg("User registered successfully")

	// Return token response
	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Data: tokenResponse,
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticates a user and returns an access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.APIResponse{data=dto.TokenResponse} "Login successful"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 403 {object} dto.ErrorResponse "Account disabled"
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

	// Process login
	tokenResponse, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Warn().Err(err).Str("email", req.Email).Msg("Login failed")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful login
	c.logger.Info().
		Str("email", req.Email).
		Msg("User logged in successfully")

	// Return token response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: tokenResponse,
	})
}

// RefreshToken handles refresh token request
// @Summary Refresh access token
// @Description Creates a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.APIResponse{data=dto.TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/refresh-token [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn().Err(err).Msg("Invalid refresh token request payload")
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Process refresh token
	tokenResponse, err := c.authService.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Refresh token failed")
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Log successful token refresh
	c.logger.Info().Msg("Token refreshed successfully")

	// Return token response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data: tokenResponse,
	})
}

// The profile-related methods have been moved to the UserController
// to eliminate duplication between /profile and /users/profile endpoints

// For backwards compatibility, these methods can be added back if needed
// but currently they are not exposed via any routes