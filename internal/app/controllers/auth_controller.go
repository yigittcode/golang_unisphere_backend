package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto" // Ensure DTO import
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

// handleError maps service errors to HTTP status codes and dto.ErrorDetail
// This is a local helper for AuthController.
func handleError(ctx *gin.Context, err error) {
	var statusCode int
	var errDetail *dto.ErrorDetail

	// Default error details
	statusCode = http.StatusInternalServerError
	errDetail = dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected internal error occurred.")
	if err != nil {
		errDetail = errDetail.WithDetails(err.Error())
	}

	// --- Specific Auth Service Error Mapping ---
	switch {
	// Validation errors from service
	case errors.Is(err, services.ErrInvalidEmail):
		statusCode = http.StatusBadRequest
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInvalidEmail, "Invalid email format")
	case errors.Is(err, services.ErrInvalidPassword):
		statusCode = http.StatusBadRequest
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInvalidPassword, "Invalid password format")
	case errors.Is(err, services.ErrInvalidStudentID):
		statusCode = http.StatusBadRequest
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInvalidStudentID, "Invalid student ID format")
	// Conflict errors
	case errors.Is(err, services.ErrEmailAlreadyExists):
		statusCode = http.StatusConflict
		errDetail = dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Email already exists")
	case errors.Is(err, services.ErrStudentIDAlreadyExists):
		statusCode = http.StatusConflict
		errDetail = dto.NewErrorDetail(dto.ErrorCodeResourceAlreadyExists, "Student ID already exists")
	// Authentication errors
	case errors.Is(err, services.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInvalidCredentials, "Invalid credentials")
	case errors.Is(err, services.ErrTokenNotFound), errors.Is(err, services.ErrTokenExpired),
		errors.Is(err, services.ErrTokenRevoked), errors.Is(err, services.ErrTokenInvalid):
		statusCode = http.StatusUnauthorized
		errDetail = dto.NewErrorDetail(dto.ErrorCodeInvalidToken, "Invalid or expired token")
	// Not found errors
	case errors.Is(err, services.ErrUserNotFound):
		statusCode = http.StatusNotFound
		errDetail = dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "User not found")
	default:
		// If the error is not specifically handled, keep the default InternalServerError
		// Log the raw error for debugging if needed (could be done in a central middleware)
		// logger.Error().Err(err).Msg("Unhandled error in auth controller")
		// Keep errDetail as initialized for generic internal server error
	}

	ctx.JSON(statusCode, dto.NewErrorResponse(errDetail))
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
	var req dto.RegisterStudentRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Use dto helper for validation errors from binding
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Map DTO to service request (Assuming service expects dto.RegisterStudentRequest)
	// If service expects a different struct, mapping is needed here.
	tokenResponse, err := c.authService.RegisterStudent(ctx, &req)
	if err != nil {
		handleError(ctx, err) // Use local error handler
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(tokenResponse, "Student registration successful")) // Use dto helper
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
	var req dto.RegisterInstructorRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Map DTO to service request (Assuming service expects dto.RegisterInstructorRequest)
	tokenResponse, err := c.authService.RegisterInstructor(ctx, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(tokenResponse, "Instructor registration successful"))
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
	var req dto.LoginRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Map DTO to service request (Assuming service expects dto.LoginRequest)
	tokenResponse, err := c.authService.Login(ctx, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(tokenResponse, "Login successful"))
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
	var req dto.RefreshTokenRequest // Use dto type
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.HandleValidationError(err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	tokenResponse, err := c.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		handleError(ctx, err)
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
// @Success 200 {object} dto.APIResponse{data=dto.UserProfile} "User profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile [get]
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		handleError(ctx, errors.New("authentication required: userID not found in context"))
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		handleError(ctx, errors.New("internal server error: invalid userID type in context"))
		return
	}

	profile, err := c.authService.GetProfile(ctx, userID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(profile, "Profile retrieved successfully"))
}

// UpdateProfilePhoto handles profile photo upload
// @Summary Update profile photo
// @Description Upload or update a user's profile photo
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Profile photo"
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.UserProfile} "Profile photo updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid file format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile/photo [post]
func (c *AuthController) UpdateProfilePhoto(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		handleError(ctx, errors.New("authentication required: userID not found in context"))
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		handleError(ctx, errors.New("internal server error: invalid userID type in context"))
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
		handleError(ctx, err)
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
// @Success 200 {object} dto.APIResponse{data=dto.UserProfile} "Profile photo deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/profile/photo [delete]
func (c *AuthController) DeleteProfilePhoto(ctx *gin.Context) {
	userIDAny, exists := ctx.Get("userID")
	if !exists {
		handleError(ctx, errors.New("authentication required: userID not found in context"))
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		handleError(ctx, errors.New("internal server error: invalid userID type in context"))
		return
	}

	profile, err := c.authService.DeleteProfilePhoto(ctx, userID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(profile, "Profile photo deleted successfully"))
}
