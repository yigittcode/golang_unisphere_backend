package controllers

import (
	"net/http"
	"strconv"
	"time"
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
// @Description Get public information about an instructor by their user ID
// @Tags instructors
// @Produce json
// @Param id path int true "Instructor User ID" Format(int64)
// @Success 200 {object} dto.APIResponse{data=dto.InstructorResponse} "Instructor retrieved successfully"
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
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert model to DTO response
	response := dto.InstructorResponse{
		ID:        instructor.ID,
		UserID:    instructor.UserID,
		Title:     instructor.Title,
	}
	
	// Add user info if available
	if instructor.User != nil {
		response.FirstName = instructor.User.FirstName
		response.LastName = instructor.User.LastName
		response.Email = instructor.User.Email
		response.CreatedAt = instructor.User.CreatedAt.Format(time.RFC3339)
		
		if instructor.User.DepartmentID != nil {
			response.DepartmentID = *instructor.User.DepartmentID
		}
	}
	
	// Add department info if available
	if instructor.Department != nil {
		response.DepartmentName = instructor.Department.Name
		
		// Add faculty info if available
		if instructor.Department.Faculty != nil {
			response.FacultyName = instructor.Department.Faculty.Name
		}
	}
	
	// Return instructor information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response, "Instructor retrieved successfully"))
}

// GetInstructorsByDepartment retrieves instructors by department ID
// @Summary Get instructors by department
// @Description Get a list of instructors belonging to a specific department
// @Tags instructors
// @Produce json
// @Param departmentId path int true "Department ID" Format(int64)
// @Success 200 {object} dto.APIResponse{data=dto.InstructorsResponse} "Instructors retrieved successfully"
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
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert models to DTOs
	instructorResponses := make([]dto.InstructorResponse, 0, len(instructors))
	for _, instructor := range instructors {
		response := dto.InstructorResponse{
			ID:        instructor.ID,
			UserID:    instructor.UserID,
			Title:     instructor.Title,
		}
		
		// Add user info if available
		if instructor.User != nil {
			response.FirstName = instructor.User.FirstName
			response.LastName = instructor.User.LastName
			response.Email = instructor.User.Email
			response.CreatedAt = instructor.User.CreatedAt.Format(time.RFC3339)
			
			if instructor.User.DepartmentID != nil {
				response.DepartmentID = *instructor.User.DepartmentID
			}
		}
		
		// Add department info if available
		if instructor.Department != nil {
			response.DepartmentName = instructor.Department.Name
		}
		
		instructorResponses = append(instructorResponses, response)
	}
	
	// Return instructor information
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.InstructorsResponse{Instructors: instructorResponses},
		"Instructors retrieved successfully",
	))
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
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert model to DTO
	response := dto.InstructorResponse{
		ID:        profile.ID,
		UserID:    profile.UserID,
		Title:     profile.Title,
	}
	
	// Add user info if available
	if profile.User != nil {
		response.FirstName = profile.User.FirstName
		response.LastName = profile.User.LastName
		response.Email = profile.User.Email
		response.CreatedAt = profile.User.CreatedAt.Format(time.RFC3339)
		
		if profile.User.DepartmentID != nil {
			response.DepartmentID = *profile.User.DepartmentID
		}
	}
	
	// Add department info if available
	if profile.Department != nil {
		response.DepartmentName = profile.Department.Name
		
		// Add faculty info if available
		if profile.Department.Faculty != nil {
			response.FacultyName = profile.Department.Faculty.Name
		}
	}
	
	// Return instructor profile
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response, "Instructor profile retrieved successfully"))
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
	successMsg := "Instructor title updated successfully"
	response := dto.SuccessResponse{Message: successMsg}
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response, successMsg))
}
