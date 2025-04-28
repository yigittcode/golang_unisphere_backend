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
// @Description Creates a new faculty with the provided information
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateFacultyRequest true "Faculty information"
// @Success 201 {object} dto.APIResponse{data=dto.FacultyResponse} "Faculty created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User does not have permission"
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
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// GetFacultyByID retrieves a faculty by ID
// @Summary Get faculty details
// @Description Retrieves detailed information about a specific faculty by its ID
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Faculty ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse{data=dto.FacultyResponse} "Faculty retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid faculty ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
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
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// GetAllFaculties retrieves all faculties
// @Summary Get all faculties
// @Description Retrieves a list of all faculties
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=[]models.Faculty} "Faculties retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties [get]
func (c *FacultyController) GetAllFaculties(ctx *gin.Context) {
	faculties, err := c.facultyService.GetAllFaculties(ctx)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
			Data:      faculties,
		Timestamp: time.Now(),
	})
}

// UpdateFaculty updates an existing faculty
// @Summary Update a faculty
// @Description Updates an existing faculty with new information
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Faculty ID" Format(int64) minimum(1)
// @Param request body dto.UpdateFacultyRequest true "Updated faculty information"
// @Success 200 {object} dto.APIResponse{data=dto.FacultyResponse} "Faculty updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User does not have permission"
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
		Data:      faculty,
		Timestamp: time.Now(),
	})
}

// DeleteFaculty deletes a faculty
// @Summary Delete a faculty
// @Description Deletes a faculty and its associated data
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Faculty ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse "Faculty deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid faculty ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User does not have permission"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
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
		Data:      nil,
		Timestamp: time.Now(),
	})
}
