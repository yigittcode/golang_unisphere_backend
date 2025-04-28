package controllers

import (
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
)

// FacultyController handles faculty-related operations
type FacultyController struct {
	facultyService services.FacultyService
}

// NewFacultyController creates a new FacultyController
func NewFacultyController(facultyService services.FacultyService) *FacultyController {
	return &FacultyController{
		facultyService: facultyService,
	}
}

// handleFacultyError is a helper function to handle common faculty error scenarios
// This controller now uses the centralized error handling middleware in middleware/error_middleware.go

// CreateFaculty handles faculty creation
// @Summary Create a new faculty
// @Description Create a new faculty with the provided data
// @Tags faculties
// @Accept json
// @Produce json
// @Param request body models.Faculty true "Faculty information"
// @Success 201 {object} dto.APIResponse "Faculty successfully created"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 409 {object} dto.ErrorResponse "Faculty already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties [post]
func (c *FacultyController) CreateFaculty(ctx *gin.Context) {
	var faculty models.Faculty
	if err := ctx.ShouldBindJSON(&faculty); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	id, err := c.facultyService.CreateFaculty(ctx, &faculty)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	faculty.ID = id
	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success:   true,
		Message:   "Faculty created successfully",
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// GetFacultyByID retrieves a faculty by ID
// @Summary Get faculty by ID
// @Description Get faculty information by ID
// @Tags faculties
// @Produce json
// @Param id path int true "Faculty ID"
// @Success 200 {object} dto.APIResponse "Faculty information retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties/{id} [get]
func (c *FacultyController) GetFacultyByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty ID")
		errorDetail = errorDetail.WithDetails("Faculty ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	faculty, err := c.facultyService.GetFacultyByID(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Faculty retrieved successfully",
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// GetAllFaculties retrieves all faculties
// @Summary Get all faculties
// @Description Get a list of all faculties
// @Tags faculties
// @Produce json
// @Success 200 {object} dto.APIResponse "Faculties retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties [get]
func (c *FacultyController) GetAllFaculties(ctx *gin.Context) {
	faculties, err := c.facultyService.GetAllFaculties(ctx)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Faculties retrieved successfully",
		Data:      faculties,
		Timestamp: time.Now(),
	})
}

// UpdateFaculty updates an existing faculty
// @Summary Update a faculty
// @Description Update a faculty with the provided data
// @Tags faculties
// @Accept json
// @Produce json
// @Param id path int true "Faculty ID"
// @Param request body models.Faculty true "Updated faculty information"
// @Success 200 {object} dto.APIResponse "Faculty successfully updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties/{id} [put]
func (c *FacultyController) UpdateFaculty(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty ID")
		errorDetail = errorDetail.WithDetails("Faculty ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	var faculty models.Faculty
	if err := ctx.ShouldBindJSON(&faculty); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Ensure the correct ID is set
	faculty.ID = id

	err = c.facultyService.UpdateFaculty(ctx, &faculty)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Faculty updated successfully",
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// DeleteFaculty deletes a faculty
// @Summary Delete a faculty
// @Description Delete a faculty by ID
// @Tags faculties
// @Produce json
// @Param id path int true "Faculty ID"
// @Success 200 {object} dto.APIResponse "Faculty successfully deleted"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
// @Failure 409 {object} dto.ErrorResponse "Cannot delete faculty with associated departments"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties/{id} [delete]
func (c *FacultyController) DeleteFaculty(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty ID")
		errorDetail = errorDetail.WithDetails("Faculty ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	err = c.facultyService.DeleteFaculty(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Faculty deleted successfully",
		Data:      nil,
		Timestamp: time.Now(),
	})
}
