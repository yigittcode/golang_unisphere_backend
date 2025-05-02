package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
)

// UserController handles user-related operations
type UserController struct {
	userService services.UserService
	fileStorage *filestorage.LocalStorage
}

// NewUserController creates a new user controller
func NewUserController(userService services.UserService, fileStorage *filestorage.LocalStorage) *UserController {
	return &UserController{
		userService: userService,
		fileStorage: fileStorage,
	}
}

// mapUserToResponse converts a user model to an extended user response DTO
func (c *UserController) mapUserToResponse(user *models.User) dto.ExtendedUserResponse {
	response := dto.ExtendedUserResponse{
		ID:                 user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Role:               string(user.RoleType),
		DepartmentID:       user.DepartmentID,
		ProfilePhotoFileID: user.ProfilePhotoFileID,
		IsActive:           user.IsActive,
	}

	// Add profile photo URL if available
	if user.ProfilePhotoFileID != nil {
		fileInfo, err := c.userService.GetFileByID(context.Background(), *user.ProfilePhotoFileID)
		if err == nil && fileInfo != nil {
			response.ProfilePhotoURL = fileInfo.FileURL
		}
	}

	return response
}

// GetUserByID retrieves user information by ID
// @Summary Get user by ID
// @Description Retrieves a specific user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse{data=dto.ExtendedUserResponse} "User retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid user ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (c *UserController) GetUserByID(ctx *gin.Context) {
	// Get user ID from URL
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid user ID")
		errorDetail = errorDetail.WithDetails("ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get user information
	user, err := c.userService.GetUserByID(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Return user information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(c.mapUserToResponse(user)))
}

// GetUsersByDepartment retrieves users by department
// @Summary Get users by department
// @Description Retrieves a list of users for a specific department with optional role filtering
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param departmentId path int true "Department ID" Format(int64) minimum(1)
// @Param role query string false "Filter by role (STUDENT, INSTRUCTOR)"
// @Success 200 {object} dto.APIResponse{data=[]dto.ExtendedUserResponse} "Users retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid department ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments/{departmentId}/users [get]
func (c *UserController) GetUsersByDepartment(ctx *gin.Context) {
	// Get department ID from URL
	departmentIdParam := ctx.Param("departmentId")
	departmentId, err := strconv.ParseInt(departmentIdParam, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department ID")
		errorDetail = errorDetail.WithDetails("Department ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get role filter if provided
	var roleFilter *string
	role := ctx.Query("role")
	if role != "" {
		roleFilter = &role
	}

	// Get users by department
	users, err := c.userService.GetUsersByDepartment(ctx, departmentId, roleFilter)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert models to DTOs
	userResponses := make([]dto.ExtendedUserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, c.mapUserToResponse(user))
	}

	// Return user information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(userResponses))
}

// GetUserProfile retrieves the profile of authenticated user
// @Summary Get user profile
// @Description Get detailed profile information for the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.ExtendedUserResponse} "User profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/profile [get]
func (c *UserController) GetUserProfile(ctx *gin.Context) {
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

	// Get user profile
	user, err := c.userService.GetUserProfile(ctx, userID)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Return user profile
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(c.mapUserToResponse(user)))
}

// UpdateUserProfile updates the user's profile information
// @Summary Update user profile
// @Description Update profile information for the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserRequest true "Profile update information"
// @Success 200 {object} dto.APIResponse{data=dto.ExtendedUserResponse} "Profile updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 409 {object} dto.ErrorResponse "Email already in use"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/profile [put]
func (c *UserController) UpdateUserProfile(ctx *gin.Context) {
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

	// Parse request body
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid profile update request")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Update user profile
	updatedUser, err := c.userService.UpdateUserProfile(ctx, userID, &req)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Return updated profile
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(c.mapUserToResponse(updatedUser)))
}

// UpdateProfilePhoto updates the user's profile photo
// @Summary Update profile photo
// @Description Update profile photo for the currently authenticated user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param photo formData file true "Profile photo to upload"
// @Success 200 {object} dto.APIResponse{data=dto.UpdateProfilePhotoResponse} "Profile photo updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid file format or missing file"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/profile/photo [post]
func (c *UserController) UpdateProfilePhoto(ctx *gin.Context) {
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

	// Get file from form
	file, err := ctx.FormFile("photo")
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid or missing file")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Update profile photo
	updatedFile, err := c.userService.UpdateProfilePhoto(ctx, userID, file)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Return success response
	response := dto.UpdateProfilePhotoResponse{
		ProfilePhotoFileID: updatedFile.ID,
	}
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// DeleteProfilePhoto deletes the user's profile photo
// @Summary Delete profile photo
// @Description Deletes the profile photo for the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=string} "Profile photo deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/profile/photo [delete]
func (c *UserController) DeleteProfilePhoto(ctx *gin.Context) {
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

	// Call auth service to delete profile photo
	err := c.userService.DeleteProfilePhoto(ctx, userID)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse("Profile photo deleted successfully"))
}

// GetUsersByFilter retrieves users based on filter criteria
// @Summary Get users by filter
// @Description Retrieves a list of users based on filter criteria with pagination
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param departmentId query int false "Filter by department ID"
// @Param role query string false "Filter by role (STUDENT, INSTRUCTOR)"
// @Param email query string false "Filter by email (partial match)"
// @Param name query string false "Filter by name (partial match on first or last name)"
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param pageSize query int false "Page size (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.UserListResponse} "Users retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid filter parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users [get]
func (c *UserController) GetUsersByFilter(ctx *gin.Context) {
	// Parse filter parameters
	var filter dto.UserFilterRequest

	// Parse pagination parameters
	pageStr := ctx.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	filter.Page = page

	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	filter.PageSize = pageSize

	// Parse department ID if provided
	departmentIDStr := ctx.Query("departmentId")
	if departmentIDStr != "" {
		departmentID, err := strconv.ParseInt(departmentIDStr, 10, 64)
		if err == nil {
			filter.DepartmentID = &departmentID
		}
	}

	// Parse role if provided
	role := ctx.Query("role")
	if role != "" {
		filter.Role = &role
	}

	// Parse email if provided
	email := ctx.Query("email")
	if email != "" {
		filter.Email = &email
	}

	// Parse name if provided
	name := ctx.Query("name")
	if name != "" {
		filter.Name = &name
	}

	// Get users by filter
	users, total, err := c.userService.GetUsersByFilter(ctx, &filter)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert models to DTOs
	userResponses := make([]dto.ExtendedUserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, c.mapUserToResponse(user))
	}

	// Create response with pagination
	response := dto.UserListResponse{
		Users: userResponses,
		PaginationInfo: dto.PaginationInfo{
			CurrentPage: filter.Page,
			TotalPages:  int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
			PageSize:    filter.PageSize,
			TotalItems:  total,
		},
	}

	// Return users
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

