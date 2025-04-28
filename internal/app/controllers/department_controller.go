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
// @Description Create a new department with the provided data
// @Tags departments
// @Accept json
// @Produce json
// @Param request body models.Department true "Department information"
// @Success 201 {object} dto.APIResponse "Department successfully created"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
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
		Success:   true,
		Message:   "Department created successfully",
		Data:      department,
		Timestamp: time.Now(),
	})
}

// GetDepartmentByID retrieves a department by ID
// @Summary Get department by ID
// @Description Get department information by ID
// @Tags departments
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} dto.APIResponse "Department information retrieved successfully"
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
		Success:   true,
		Message:   "Department retrieved successfully",
		Data:      department,
		Timestamp: time.Now(),
	})
}

// GetAllDepartments retrieves all departments
// @Summary Get all departments
// @Description Get a list of all departments
// @Tags departments
// @Produce json
// @Success 200 {object} dto.APIResponse "Departments retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /departments [get]
func (c *DepartmentController) GetAllDepartments(ctx *gin.Context) {
	departments, err := c.departmentService.GetAllDepartments(ctx)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Departments retrieved successfully",
		Data:      departments,
		Timestamp: time.Now(),
	})
}

// GetDepartmentsByFacultyID retrieves all departments for a faculty
// @Summary Get departments by faculty ID
// @Description Get a list of departments for a specific faculty
// @Tags departments
// @Produce json
// @Param facultyId path int true "Faculty ID"
// @Success 200 {object} dto.APIResponse "Departments retrieved successfully"
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
		Success:   true,
		Message:   "Departments retrieved successfully",
		Data:      departments,
		Timestamp: time.Now(),
	})
}

// UpdateDepartment updates an existing department
// @Summary Update a department
// @Description Update a department with the provided data
// @Tags departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Param request body models.Department true "Updated department information"
// @Success 200 {object} dto.APIResponse "Department successfully updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
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
		Success:   true,
		Message:   "Department updated successfully",
		Data:      department,
		Timestamp: time.Now(),
	})
}

// DeleteDepartment deletes a department
// @Summary Delete a department
// @Description Delete a department by ID
// @Tags departments
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} dto.APIResponse "Department successfully deleted"
// @Failure 404 {object} dto.ErrorResponse "Department not found"
// @Failure 409 {object} dto.ErrorResponse "Cannot delete department with associated data"
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

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Department deleted successfully",
		Data:      nil,
		Timestamp: time.Now(),
	})
}
