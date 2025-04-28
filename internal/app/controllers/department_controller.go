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

// DepartmentController handles department-related operations
type DepartmentController struct {
	departmentService services.DepartmentService
}

// NewDepartmentController creates a new DepartmentController
func NewDepartmentController(departmentService services.DepartmentService) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
	}
}

// handleDepartmentError is a helper function to handle common department error scenarios
// This controller now uses the centralized error handling middleware in middleware/error_middleware.go

// CreateDepartment handles department creation
// @Summary Create a new department
// @Description Creates a new department with the provided information
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateDepartmentRequest true "Department information"
// @Success 201 {object} dto.APIResponse{data=dto.DepartmentResponse} "Department created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request data"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User does not have permission"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
// @Failure 409 {object} dto.ErrorResponse "Department already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments [post]
func (c *DepartmentController) CreateDepartment(ctx *gin.Context) {
	var department models.Department
	if err := ctx.ShouldBindJSON(&department); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	err := c.departmentService.CreateDepartment(ctx, &department)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Data:      department,
		Timestamp: time.Now(),
	})
}

// GetDepartmentByID retrieves a department by ID
// @Summary Get department by ID
// @Description Retrieves a specific department by its ID
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} dto.APIResponse{data=models.Department} "Department retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid department ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments/{id} [get]
func (c *DepartmentController) GetDepartmentByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department ID")
		errorDetail = errorDetail.WithDetails("Department ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	department, err := c.departmentService.GetDepartmentByID(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data:      department,
		Timestamp: time.Now(),
	})
}

// GetAllDepartments retrieves all departments
// @Summary Get all departments
// @Description Retrieves a list of all departments
// @Tags departments
// @Accept json
// @Produce json
// @Param facultyId query int false "Filter by faculty ID"
// @Success 200 {object} dto.APIResponse{data=[]models.Department} "Departments retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments [get]
func (c *DepartmentController) GetAllDepartments(ctx *gin.Context) {
	departments, err := c.departmentService.GetAllDepartments(ctx)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data:      departments,
		Timestamp: time.Now(),
	})
}

// GetDepartmentsByFacultyID retrieves all departments for a faculty
// @Summary List faculty departments
// @Description Retrieves a list of all departments belonging to a specific faculty
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param facultyId path int true "Faculty ID" Format(int64) minimum(1)
// @Success 200 {object} dto.APIResponse{data=[]dto.DepartmentResponse} "Faculty departments retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid faculty ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 404 {object} dto.ErrorResponse "Faculty not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /faculty-departments/{facultyId} [get]
func (c *DepartmentController) GetDepartmentsByFacultyID(ctx *gin.Context) {
	facultyIDStr := ctx.Param("facultyId")
	facultyID, err := strconv.ParseInt(facultyIDStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid faculty ID")
		errorDetail = errorDetail.WithDetails("Faculty ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	departments, err := c.departmentService.GetDepartmentsByFacultyID(ctx, facultyID)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data:      departments,
		Timestamp: time.Now(),
	})
}

// UpdateDepartment updates an existing department
// @Summary Update a department
// @Description Updates an existing department with the provided information
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Param request body dto.UpdateDepartmentRequest true "Updated department information"
// @Success 200 {object} dto.APIResponse{data=models.Department} "Department updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments/{id} [put]
func (c *DepartmentController) UpdateDepartment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department ID")
		errorDetail = errorDetail.WithDetails("Department ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	var department models.Department
	if err := ctx.ShouldBindJSON(&department); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Ensure the correct ID is set
	department.ID = id

	err = c.departmentService.UpdateDepartment(ctx, &department)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Data:      department,
		Timestamp: time.Now(),
	})
}

// DeleteDepartment deletes a department
// @Summary Delete a department
// @Description Deletes an existing department by its ID
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 204 "Department deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid department ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments/{id} [delete]
func (c *DepartmentController) DeleteDepartment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid department ID")
		errorDetail = errorDetail.WithDetails("Department ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	err = c.departmentService.DeleteDepartment(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, dto.APIResponse{
		Data:      nil,
		Timestamp: time.Now(),
	})
}
