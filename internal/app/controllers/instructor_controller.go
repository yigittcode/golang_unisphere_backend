package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
)

// InstructorController handles instructor related operations
type InstructorController struct {
	instructorService services.InstructorService
}

// NewInstructorController creates a new instructor controller
func NewInstructorController(instructorService services.InstructorService) *InstructorController {
	return &InstructorController{
		instructorService: instructorService,
	}
}

// handleInstructorError is a helper function to handle common instructor error scenarios
// This controller now uses the centralized error handling middleware in middleware/error_middleware.go

// GetInstructorByID retrieves instructor information by ID
// @Summary Get instructor by ID
// @Description Retrieves a specific instructor by its ID
// @Tags instructors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Instructor ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse{data=dto.InstructorResponse} "Instructor retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid instructor ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "Instructor not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /instructors/{id} [get]
func (c *InstructorController) GetInstructorByID(ctx *gin.Context) {
	// Get instructor ID from URL
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid instructor ID")
		errorDetail = errorDetail.WithDetails("ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get instructor information
	instructor, err := c.instructorService.GetInstructorByID(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert model to DTO response
	response := dto.InstructorResponse{
		UserResponse: dto.UserResponse{
			ID:           instructor.User.ID,
			Email:        instructor.User.Email,
			FirstName:    instructor.User.FirstName,
			LastName:     instructor.User.LastName,
			Role:         string(instructor.User.RoleType),
			DepartmentID: instructor.User.DepartmentID,
		},
		Title: instructor.Title,
	}

	// Return instructor information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// GetInstructorsByDepartment retrieves instructors by department
// @Summary Get instructors by department
// @Description Retrieves a list of instructors for a specific department
// @Tags instructors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param departmentId path int true "Department ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse{data=[]dto.InstructorResponse} "Instructors retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid department ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/department-instructors/{departmentId} [get]
func (c *InstructorController) GetInstructorsByDepartment(ctx *gin.Context) {
	// Get department ID from URL
	departmentIdParam := ctx.Param("departmentId")
	departmentId, err := strconv.ParseInt(departmentIdParam, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department ID")
		errorDetail = errorDetail.WithDetails("Department ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get instructors by department
	instructors, err := c.instructorService.GetInstructorsByDepartment(ctx, departmentId)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert models to DTOs
	instructorResponses := make([]dto.InstructorResponse, 0, len(instructors))
	for _, instructor := range instructors {
		if instructor.User != nil {
			response := dto.InstructorResponse{
				UserResponse: dto.UserResponse{
					ID:           instructor.User.ID,
					Email:        instructor.User.Email,
					FirstName:    instructor.User.FirstName,
					LastName:     instructor.User.LastName,
					Role:         string(instructor.User.RoleType),
					DepartmentID: instructor.User.DepartmentID,
				},
				Title: instructor.Title,
			}
			instructorResponses = append(instructorResponses, response)
		}
	}

	// Return instructor information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(instructorResponses))
}

// GetInstructorProfile retrieves the profile of authenticated instructor
// @Summary Get instructor profile
// @Description Get detailed profile information for the currently authenticated instructor (Requires instructor role)
// @Tags instructors
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.InstructorResponse} "Instructor profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User is not an instructor or token invalid"
// @Failure 404 {object} dto.ErrorResponse "Instructor not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/instructors/profile [get]
func (c *InstructorController) GetInstructorProfile(ctx *gin.Context) {
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

	// Get instructor profile
	profile, err := c.instructorService.GetInstructorWithDetails(ctx, userID)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert model to DTO
	response := dto.InstructorResponse{
		UserResponse: dto.UserResponse{
			ID:           profile.User.ID,
			Email:        profile.User.Email,
			FirstName:    profile.User.FirstName,
			LastName:     profile.User.LastName,
			Role:         string(profile.User.RoleType),
			DepartmentID: profile.User.DepartmentID,
		},
		Title: profile.Title,
	}

	// Return instructor profile
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// UpdateTitle updates the title of an instructor
// @Summary Update instructor title
// @Description Update the academic title for the currently authenticated instructor (Requires instructor role)
// @Tags instructors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateTitleRequest true "New title information"
// @Success 200 {object} dto.APIResponse "Instructor title updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error (e.g., empty title)"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User is not an instructor or token invalid"
// @Failure 404 {object} dto.ErrorResponse "Instructor not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/instructors/title [put]
func (c *InstructorController) UpdateTitle(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Authentication required")
		errorDetail = errorDetail.WithDetails("User ID not found in request context")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert user ID to int64
	userIDInt64, ok := userIDInterface.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid user ID format")
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Parse request body
	var req dto.UpdateTitleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid title update request")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Update instructor title
	err := c.instructorService.UpdateTitle(ctx, userIDInt64, req.Title)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Return success response
	response := dto.SuccessResponse{Message: "Instructor title updated successfully"}
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}
