package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// PastExamController handles past exam related operations
type PastExamController struct {
	pastExamService services.PastExamService
	fileStorage     *filestorage.LocalStorage
}

// NewPastExamController creates a new PastExamController
func NewPastExamController(pastExamService services.PastExamService, fileStorage *filestorage.LocalStorage) *PastExamController {
	return &PastExamController{
		pastExamService: pastExamService,
		fileStorage:     fileStorage,
	}
}

// This controller now uses the centralized error handling middleware

// toPastExamResponse converts a PastExam model to a PastExamResponse DTO
func toPastExamResponse(exam *models.PastExam) dto.PastExamResponse {
	return dto.PastExamResponse{
		ID:           exam.ID,
		CourseCode:   exam.CourseCode,
		Year:         exam.Year,
		Term:         exam.Term,
		FileID:       exam.FileID,
		DepartmentID: exam.DepartmentID,
		InstructorID: exam.InstructorID,
	}
}

// GetAllPastExams handles retrieving all past exams with optional filtering
// @Summary Get all past exams
// @Description Retrieves a list of past exams with optional filtering and pagination
// @Tags past-exams
// @Accept json
// @Produce json
// @Param facultyId query int false "Filter by faculty ID"
// @Param departmentId query int false "Filter by department ID"
// @Param courseCode query string false "Filter by course code"
// @Param year query int false "Filter by year"
// @Param term query string false "Filter by term (FALL, SPRING, SUMMER)"
// @Param sortBy query string false "Sort field (year, term, courseCode, title, departmentName, facultyName, instructorName, createdAt, updatedAt)"
// @Param sortOrder query string false "Sort order (ASC, DESC)"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 10)"
// @Success 200 {object} dto.APIResponse{data=[]models.PastExam,pagination=dto.PaginationInfo} "Past exams retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams [get]
func (c *PastExamController) GetAllPastExams(ctx *gin.Context) {
	// Parse pagination parameters (0-based)
	page, err := strconv.Atoi(ctx.DefaultQuery("page", strconv.Itoa(helpers.DefaultPage)))
	if err != nil || page < 0 {
		page = helpers.DefaultPage
	}
	// page = page + 1 // REMOVED 1-based conversion

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("size", strconv.Itoa(helpers.DefaultPageSize)))
	if err != nil || pageSize <= 0 || pageSize > helpers.MaxPageSize {
		pageSize = helpers.DefaultPageSize
	}

	// Parse filters
	filters := make(map[string]interface{})

	// Add facultyId filter if provided
	if facultyIDStr := ctx.Query("facultyId"); facultyIDStr != "" {
		if facultyID, err := strconv.ParseInt(facultyIDStr, 10, 64); err == nil {
			filters["facultyId"] = facultyID
		}
	}

	// Add departmentId filter if provided
	if deptIDStr := ctx.Query("departmentId"); deptIDStr != "" {
		if deptID, err := strconv.ParseInt(deptIDStr, 10, 64); err == nil {
			filters["departmentId"] = deptID
		}
	}

	// Add courseCode filter if provided
	if courseCode := ctx.Query("courseCode"); courseCode != "" {
		filters["courseCode"] = courseCode
	}

	// Add year filter if provided
	if yearStr := ctx.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filters["year"] = year
		}
	}

	// Add term filter if provided
	if term := ctx.Query("term"); term != "" {
		filters["term"] = term
	}

	// Add sorting parameters if provided
	if sortBy := ctx.Query("sortBy"); sortBy != "" {
		filters["sortBy"] = sortBy
	}

	if sortOrder := ctx.Query("sortOrder"); sortOrder != "" {
		filters["sortOrder"] = sortOrder
	}

	// Get past exams from service
	filter := &dto.PastExamFilterRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// Add filters if provided
	if deptID, ok := filters["departmentId"].(int64); ok {
		filter.DepartmentID = &deptID
	}
	if courseCode, ok := filters["courseCode"].(string); ok {
		filter.CourseCode = &courseCode
	}
	if year, ok := filters["year"].(int); ok {
		filter.Year = &year
	}
	if term, ok := filters["term"].(string); ok {
		filter.Term = &term
	}

	response, err := c.pastExamService.GetAllExams(ctx, filter)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// GetPastExamByID handles retrieving a specific past exam by ID
// @Summary Get past exam by ID
// @Description Retrieves a specific past exam by its ID
// @Tags past-exams
// @Accept json
// @Produce json
// @Param id path int true "Past exam ID"
// @Success 200 {object} dto.APIResponse{data=models.PastExam} "Past exam retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid past exam ID"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams/{id} [get]
func (c *PastExamController) GetPastExamByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get exam by ID
	exam, err := c.pastExamService.GetExamByID(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(exam))
}

// CreatePastExam handles creating a new past exam
// @Summary Create a new past exam
// @Description Creates a new past exam with the provided information
// @Tags past-exams
// @Accept multipart/form-data
// @Produce json
// @Param year formData int true "Year of the exam"
// @Param term formData string true "Term of the exam (FALL, SPRING, SUMMER)"
// @Param departmentId formData int true "Department ID"
// @Param courseCode formData string true "Course code"
// @Param title formData string true "Exam title"
// @Param instructorId formData int false "Instructor ID"
// @Param file formData file true "Exam file"
// @Success 201 {object} dto.APIResponse{data=models.PastExam} "Past exam created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams [post]
func (c *PastExamController) CreatePastExam(ctx *gin.Context) {
	// Parse form data instead of JSON
	var req dto.CreatePastExamRequest
	// Bind form values to the struct (note: file needs separate handling)
	if err := ctx.ShouldBind(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid form data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get user ID from context
	_, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid or missing file")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to create past exam
	createdExam, err := c.pastExamService.CreateExam(ctx, &req, file)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(createdExam))
}

// UpdatePastExam handles updating an existing past exam
// @Summary Update a past exam
// @Description Updates an existing past exam with the provided information
// @Tags past-exams
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Past exam ID"
// @Param year formData int false "Year of the exam"
// @Param term formData string false "Term of the exam (FALL, SPRING, SUMMER)"
// @Param departmentId formData int false "Department ID"
// @Param courseCode formData string false "Course code"
// @Param title formData string false "Exam title"
// @Param instructorId formData int false "Instructor ID"
// @Param file formData file false "Exam file"
// @Success 200 {object} dto.APIResponse{data=models.PastExam} "Past exam updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams/{id} [put]
func (c *PastExamController) UpdatePastExam(ctx *gin.Context) {
	// Get exam ID from URL
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Parse form data instead of JSON
	var req dto.UpdatePastExamRequest
	// Bind form values to the struct (note: file needs separate handling)
	if err := ctx.ShouldBind(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid form data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get user ID from context
	_, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to update past exam
	updatedExam, err := c.pastExamService.UpdateExam(ctx, id, &req)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(updatedExam))
}

// DeletePastExam handles deleting a past exam
// @Summary Delete a past exam
// @Description Deletes an existing past exam by its ID
// @Tags past-exams
// @Accept json
// @Produce json
// @Param id path int true "Past exam ID"
// @Success 204 "Past exam deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid past exam ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams/{id} [delete]
func (c *PastExamController) DeletePastExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get user ID from context
	_, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to delete past exam
	err = c.pastExamService.DeleteExam(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrPastExamNotFound) {
			// If exam doesn't exist, just return the error
			middleware.HandleAPIError(ctx, err)
			return
		}
		// For other errors, log and continue with deletion attempt
		logger.Warn().Err(err).Int64("examID", id).Msg("Error getting exam details before deletion")
	}

	// Delete associated files
	if err := c.fileStorage.DeleteFile(fmt.Sprintf("past_exam_%d", id)); err != nil {
		logger.Error().Err(err).Int64("examId", id).Msg("Failed to delete files from storage")
	}

	ctx.JSON(http.StatusNoContent, nil)
}
