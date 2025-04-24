package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
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
	if errors.Is(err, services.ErrInstructorNotFound) {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Instructor not found")
		errorDetail = errorDetail.WithDetails("The requested instructor does not exist")
		ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(errorDetail))
		return
	} else if errors.Is(err, services.ErrUnauthorized) {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Unauthorized operation")
		errorDetail = errorDetail.WithDetails("User is not an instructor")
		ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(errorDetail))
		return
	}

	// Default case for unexpected errors
	errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An error occurred while processing your request")
	errorDetail = errorDetail.WithDetails(err.Error())
	ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
}

// GetInstructorByID retrieves instructor information by ID
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
