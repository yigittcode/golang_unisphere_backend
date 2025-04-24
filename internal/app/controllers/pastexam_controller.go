package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
)

// PastExamController handles past exam related operations
type PastExamController struct {
	pastExamService *services.PastExamService
}

// NewPastExamController creates a new PastExamController
func NewPastExamController(pastExamService *services.PastExamService) *PastExamController {
	return &PastExamController{
		pastExamService: pastExamService,
	}
}

// handlePastExamError is a helper function to handle common past exam error scenarios
func handlePastExamError(ctx *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorCode := dto.ErrorCodeInternalServer
	errorMessage := "An unexpected error occurred"
	errorDetails := err.Error()

	// Handle specific errors
	switch {
	case errors.Is(err, services.ErrPastExamNotFound):
		statusCode = http.StatusNotFound
		errorCode = dto.ErrorCodeResourceNotFound
		errorMessage = "Past exam not found"
		errorDetails = "The requested past exam does not exist"
	case errors.Is(err, services.ErrInstructorOnly):
		statusCode = http.StatusForbidden
		errorCode = dto.ErrorCodeUnauthorized
		errorMessage = "Instructors only"
		errorDetails = "Only instructors can create past exams"
	case errors.Is(err, services.ErrPermissionDenied):
		statusCode = http.StatusForbidden
		errorCode = dto.ErrorCodeUnauthorized
		errorMessage = "Permission denied"
		errorDetails = "You don't have permission to perform this action"
	}

	errorDetail := dto.NewErrorDetail(errorCode, errorMessage)
	errorDetail = errorDetail.WithDetails(errorDetails)
	ctx.JSON(statusCode, dto.NewErrorResponse(errorDetail))
}

// GetAllPastExams retrieves all past exams with pagination and filtering
// @Summary Get all past exams
// @Description Get a list of all past exams with optional filtering and pagination
// @Tags pastexams
// @Produce json
// @Param facultyId query int false "Filter by faculty ID"
// @Param departmentId query int false "Filter by department ID"
// @Param courseCode query string false "Filter by course code"
// @Param year query int false "Filter by year"
// @Param term query string false "Filter by term (FALL or SPRING)"
// @Param sortBy query string false "Sort field (default: createdAt)"
// @Param sortOrder query string false "Sort order (ASC or DESC, default: DESC)"
// @Param page query int false "Page number (0-based, default: 0)"
// @Param size query int false "Page size (default: 10)"
// @Success 200 {object} dto.APIResponse "Past exams retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams [get]
func (c *PastExamController) GetAllPastExams(ctx *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	if err != nil || page < 0 {
		page = 0
	}
	// API is 0-based, but service expects 1-based pagination
	page += 1

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
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
	pastExams, totalItems, err := c.pastExamService.GetAllPastExams(ctx, page, pageSize, filters)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Convert to response DTOs
	examResponses := make([]dto.PastExamResponse, 0, len(pastExams))
	for _, exam := range pastExams {
		examResponses = append(examResponses, dto.FromPastExam(&exam))
	}

	// Create response with pagination info
	totalPages := (totalItems + pageSize - 1) / pageSize
	response := dto.PastExamListResponse{
		Exams: examResponses,
		Pagination: dto.PaginationInfo{
			CurrentPage: page - 1, // convert back to 0-based for API
			TotalPages:  totalPages,
			PageSize:    pageSize,
			TotalItems:  totalItems,
		},
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Past exams retrieved successfully",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// GetPastExamByID retrieves a past exam by ID
// @Summary Get past exam by ID
// @Description Get past exam information by ID
// @Tags pastexams
// @Produce json
// @Param id path int true "Past Exam ID"
// @Success 200 {object} dto.APIResponse "Past exam information retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams/{id} [get]
func (c *PastExamController) GetPastExamByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	pastExam, err := c.pastExamService.GetPastExamByID(ctx, id)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	response := dto.FromPastExam(pastExam)
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Past exam retrieved successfully",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// CreatePastExam handles past exam creation
// @Summary Create a new past exam
// @Description Create a new past exam with the provided data
// @Tags pastexams
// @Accept json
// @Produce json
// @Param request body dto.CreatePastExamRequest true "Past exam information"
// @Success 201 {object} dto.APIResponse "Past exam successfully created"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 403 {object} dto.ErrorResponse "Permission denied or not an instructor"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams [post]
func (c *PastExamController) CreatePastExam(ctx *gin.Context) {
	// Get validated request from context
	validatedObj, exists := ctx.Get("validatedBody")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	createRequest := validatedObj.(*dto.CreatePastExamRequest)

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert request to model
	pastExam := &models.PastExam{
		Year:         createRequest.Year,
		Term:         models.Term(createRequest.Term),
		DepartmentID: createRequest.DepartmentID,
		CourseCode:   createRequest.CourseCode,
		Title:        createRequest.Title,
		Content:      createRequest.Content,
		FileURL:      createRequest.FileURL,
		// InstructorID will be set in the service based on user's instructor record
	}

	// Call service to create past exam
	id, err := c.pastExamService.CreatePastExam(ctx, pastExam, userID.(int64))
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Get the created past exam with all details
	createdExam, err := c.pastExamService.GetPastExamByID(ctx, id)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	response := dto.FromPastExam(createdExam)
	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success:   true,
		Message:   "Past exam created successfully",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// UpdatePastExam updates an existing past exam
// @Summary Update a past exam
// @Description Update a past exam with the provided data
// @Tags pastexams
// @Accept json
// @Produce json
// @Param id path int true "Past Exam ID"
// @Param request body dto.UpdatePastExamRequest true "Updated past exam information"
// @Success 200 {object} dto.APIResponse "Past exam successfully updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 403 {object} dto.ErrorResponse "Permission denied or not an instructor"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams/{id} [put]
func (c *PastExamController) UpdatePastExam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get validated request from context
	validatedObj, exists := ctx.Get("validatedBody")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Validation failed")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	updateRequest := validatedObj.(*dto.UpdatePastExamRequest)

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert request to model
	pastExam := &models.PastExam{
		ID:           id,
		Year:         updateRequest.Year,
		Term:         models.Term(updateRequest.Term),
		DepartmentID: updateRequest.DepartmentID,
		CourseCode:   updateRequest.CourseCode,
		Title:        updateRequest.Title,
		Content:      updateRequest.Content,
		FileURL:      updateRequest.FileURL,
	}

	// Call service to update past exam
	err = c.pastExamService.UpdatePastExam(ctx, pastExam, userID.(int64))
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Get the updated past exam with all details
	updatedExam, err := c.pastExamService.GetPastExamByID(ctx, id)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	response := dto.FromPastExam(updatedExam)
	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Past exam updated successfully",
		Data:      response,
		Timestamp: time.Now(),
	})
}

// DeletePastExam deletes a past exam
// @Summary Delete a past exam
// @Description Delete a past exam by ID
// @Tags pastexams
// @Produce json
// @Param id path int true "Past Exam ID"
// @Success 200 {object} dto.APIResponse "Past exam successfully deleted"
// @Failure 403 {object} dto.ErrorResponse "Permission denied or not an instructor"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams/{id} [delete]
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
	userID, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to delete past exam
	err = c.pastExamService.DeletePastExam(ctx, id, userID.(int64))
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Past exam deleted successfully",
		Data:      nil,
		Timestamp: time.Now(),
	})
}
