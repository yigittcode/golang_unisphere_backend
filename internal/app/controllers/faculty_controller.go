package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/helpers"
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
	var req dto.CreateFacultyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert DTO to model
	faculty := &models.Faculty{
		Name: req.Name,
		Code: req.Code,
	}

	id, err := c.facultyService.CreateFaculty(ctx, faculty)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	faculty.ID = id

	// Create response
	response := dto.FacultyResponse{
		ID:   faculty.ID,
		Name: faculty.Name,
		Code: faculty.Code,
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(response))
}

// GetFacultyByID retrieves a faculty by ID
// @Summary Get faculty details
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

	// Create response
	response := dto.FacultyResponse{
		ID:   faculty.ID,
		Name: faculty.Name,
		Code: faculty.Code,
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// GetAllFaculties retrieves all faculties
// @Summary Get all faculties
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param size query int false "Page size" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.FacultyListResponse} "Faculties retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculties [get]
func (c *FacultyController) GetAllFaculties(ctx *gin.Context) {
	// Parse pagination parameters
	page, size := helpers.ParsePaginationParams(ctx)

	faculties, err := c.facultyService.GetAllFaculties(ctx)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Convert to response DTOs
	var facultyResponses []dto.FacultyResponse
	for _, faculty := range faculties {
		facultyResponses = append(facultyResponses, dto.FacultyResponse{
			ID:   faculty.ID,
			Name: faculty.Name,
			Code: faculty.Code,
		})
	}

	// Calculate pagination values
	totalItems := int64(len(facultyResponses))
	paginationInfo := helpers.NewPaginationInfo(totalItems, page, size)

	// Apply pagination to results if needed
	start, end := helpers.CalculateSliceIndices(page, size, len(facultyResponses))
	if start < len(facultyResponses) {
		if end > len(facultyResponses) {
			end = len(facultyResponses)
		}
		facultyResponses = facultyResponses[start:end]
	} else {
		facultyResponses = []dto.FacultyResponse{}
	}

	// Create response
	response := dto.FacultyListResponse{
		Faculties:      facultyResponses,
		PaginationInfo: paginationInfo,
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// UpdateFaculty updates an existing faculty
// @Summary Update a faculty
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

	var req dto.UpdateFacultyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert DTO to model
	faculty := &models.Faculty{
		ID:   id,
		Name: req.Name,
		Code: req.Code,
	}

	err = c.facultyService.UpdateFaculty(ctx, faculty)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	// Create response
	response := dto.FacultyResponse{
		ID:   faculty.ID,
		Name: faculty.Name,
		Code: faculty.Code,
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// DeleteFaculty deletes a faculty
// @Summary Delete a faculty
// @Tags faculties
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Faculty ID" Format(int64) minimum(1)
// @Success 204 "Faculty deleted successfully"
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

	ctx.JSON(http.StatusNoContent, nil)
}
