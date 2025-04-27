package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// InstructorController handles instructor related operations
type InstructorController struct {
	instructorService *services.InstructorService
}

// NewInstructorController creates a new instructor controller
func NewInstructorController(instructorService *services.InstructorService) *InstructorController {
	return &InstructorController{
		instructorService: instructorService,
	}
}

// handleInstructorError is a helper function to handle common instructor error scenarios
func handleInstructorError(ctx *gin.Context, err error) {
	if errors.Is(err, apperrors.ErrUserNotFound) {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Instructor not found")
		errorDetail = errorDetail.WithDetails("The requested instructor does not exist")
		ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(errorDetail))
		return
	} else if errors.Is(err, apperrors.ErrPermissionDenied) {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Unauthorized operation")
		errorDetail = errorDetail.WithDetails("User is not an instructor")
		ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(errorDetail))
		return
	} else if errors.Is(err, apperrors.ErrValidationFailed) {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Default case for unexpected errors
	errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An error occurred while processing your request")
	errorDetail = errorDetail.WithDetails(err.Error())
	ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
}

// GetInstructorByID retrieves instructor information by ID
// @Summary Get instructor by ID
// @Description Get public information about an instructor by their user ID
// @Tags instructors
// @Produce json
// @Param id path int true "Instructor User ID" Format(int64)
// @Success 200 {object} dto.APIResponse{data=models.Instructor} "Instructor retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid Instructor ID format"
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
		handleInstructorError(ctx, err)
		return
	}

	// Return instructor information
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Instructor retrieved successfully",
		Data:      instructor,
		Timestamp: time.Now(),
	})
}

// GetInstructorsByDepartment retrieves instructors by department ID
// @Summary Get instructors by department
// @Description Get a list of instructors belonging to a specific department
// @Tags instructors
// @Produce json
// @Param departmentId path int true "Department ID" Format(int64)
// @Success 200 {object} dto.APIResponse{data=[]models.Instructor} "Instructors retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid Department ID format"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /department-instructors/{departmentId} [get] // Note: Path defined in routes.go
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
		handleInstructorError(ctx, err)
		return
	}

	// Return instructor information
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Instructors retrieved successfully",
		Data:      instructors,
		Timestamp: time.Now(),
	})
}

// GetInstructorProfile retrieves the profile of authenticated instructor
// @Summary Get instructor profile
// @Description Get detailed profile information for the currently authenticated instructor (Requires instructor role)
// @Tags instructors
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=models.Instructor} "Instructor profile retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User is not an instructor or token invalid"
// @Failure 404 {object} dto.ErrorResponse "Instructor not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /instructors/profile [get]
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
		handleInstructorError(ctx, err)
		return
	}

	// Return instructor profile
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Instructor profile retrieved successfully",
		Data:      profile,
		Timestamp: time.Now(),
	})
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
// @Router /instructors/title [put]
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
	// Define UpdateTitleRequest DTO struct (Consider moving to dto package)
	type UpdateTitleRequest struct {
		Title string `json:"title" binding:"required" example:"Associate Professor"` // New academic title for the instructor
	}
	var req UpdateTitleRequest // Use the local definition
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid title update request")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Update instructor title
	err := c.instructorService.UpdateInstructorTitle(ctx, userIDInt64, req.Title)
	if err != nil {
		handleInstructorError(ctx, err)
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Instructor title updated successfully",
		Timestamp: time.Now(),
	})
}
