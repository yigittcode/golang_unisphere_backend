package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// PastExamController handles past exam related operations
type PastExamController struct {
	pastExamService *services.PastExamService
	fileStorage     *filestorage.LocalStorage
}

// NewPastExamController creates a new PastExamController
func NewPastExamController(pastExamService *services.PastExamService, fileStorage *filestorage.LocalStorage) *PastExamController {
	return &PastExamController{
		pastExamService: pastExamService,
		fileStorage:     fileStorage,
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
// @Description Get a list of all past exams. Supports filtering by faculty, department, course code, year, term and sorting by various fields.
// @Tags pastexams
// @Produce json
// @Param facultyId query int false "Filter by faculty ID" Format(int64) example(1)
// @Param departmentId query int false "Filter by department ID" Format(int64) example(1)
// @Param courseCode query string false "Filter by course code (case-insensitive, partial match)" example(CENG)
// @Param year query int false "Filter by exact year" example(2023)
// @Param term query string false "Filter by term" Enums(FALL, SPRING) example(FALL)
// @Param sortBy query string false "Sort field (year, term, courseCode, title, createdAt, updatedAt)" default(createdAt)
// @Param sortOrder query string false "Sort order" Enums(ASC, DESC) default(DESC)
// @Param page query int false "Page number for pagination (0-based)" default(0) minimum(0)
// @Param size query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.PastExamListResponse} "Past exams retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameter format or value"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams [get]
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

	// Get past exams from service (page is 0-based)
	pastExams, paginationInfo, err := c.pastExamService.GetAllPastExams(ctx, page, pageSize, filters)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Convert to response DTOs
	examResponses := make([]dto.PastExamResponse, 0, len(pastExams))
	for _, exam := range pastExams {
		examResponses = append(examResponses, dto.FromPastExam(&exam))
	}

	// Create response with pagination info from service
	response := dto.PastExamListResponse{
		Exams: examResponses,
		// Pagination info is now directly from the service response (which uses the helper)
		Pagination: paginationInfo,
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response, "Past exams retrieved successfully"))
}

// GetPastExamByID retrieves a past exam by ID
// @Summary Get past exam by ID
// @Description Get detailed information for a specific past exam by its ID.
// @Tags pastexams
// @Produce json
// @Param id path int true "Past Exam ID" Format(int64) example(1)
// @Success 200 {object} dto.APIResponse{data=dto.PastExamResponse} "Past exam information retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid Past Exam ID format"
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
// @Summary Create a new past exam (Instructor only)
// @Description Create a new past exam with the provided data and optional file upload.
// @Tags pastexams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param year formData int true "Year" example(2023)
// @Param term formData string true "Term (FALL or SPRING)" example(FALL)
// @Param departmentId formData int true "Department ID" example(1)
// @Param courseCode formData string true "Course Code" example(CENG301)
// @Param title formData string true "Title" example("Midterm Exam")
// @Param content formData string true "Content" example("Exam content details...")
// @Param files formData file false "Exam files (PDF, image, etc.)" collectionFormat(multi)
// @Success 201 {object} dto.APIResponse{data=dto.PastExamResponse} "Past exam successfully created"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User is not an instructor"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams [post]
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
	userID, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Invalid userID type")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert request to model
	pastExam := &models.PastExam{
		Year:         req.Year,
		Term:         models.Term(req.Term),
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
	}

	// Call service to create past exam
	id, err := c.pastExamService.CreatePastExam(ctx, pastExam, userIDInt)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Handle multiple file uploads
	form, err := ctx.MultipartForm()
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		logger.Error().Err(err).Msg("Error retrieving multipart form")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Error processing file upload")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// If we have files, process them
	uploadedFiles := []*models.File{}
	if form != nil && form.File != nil {
		files := form.File["files"]
		for _, fileHeader := range files {
			// Save the file
			savedFilePath, err := c.fileStorage.SaveFile(fileHeader)
			if err != nil {
				logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error saving uploaded file")
				// Continue with next file if one fails
				continue
			}

			// Create file record in database
			savedFile := &models.File{
				FileName:     fileHeader.Filename,
				FilePath:     savedFilePath,
				FileURL:      c.fileStorage.GetFileURL(savedFilePath),
				FileSize:     fileHeader.Size,
				FileType:     fileHeader.Header.Get("Content-Type"),
				ResourceType: models.FileTypePastExam,
				ResourceID:   id,
				UploadedBy:   userIDInt,
			}

			// Add file to uploaded files collection
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	// If we have any files, associate them with the past exam
	for _, file := range uploadedFiles {
		fileID, err := c.pastExamService.AddFileToPastExam(ctx, id, file)
		if err != nil {
			logger.Error().Err(err).Int64("examId", id).Str("filename", file.FileName).Msg("Error attaching file to past exam")
			// Continue with next file if one fails
		}
		file.ID = fileID
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

// UpdatePastExam handles past exam updates
// @Summary Update a past exam (Instructor only)
// @Description Update an existing past exam with the provided data and optional new file upload.
// @Tags pastexams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Past Exam ID" Format(int64) example(1)
// @Param year formData int true "Year" example(2023)
// @Param term formData string true "Term (FALL or SPRING)" example(FALL)
// @Param departmentId formData int true "Department ID" example(1)
// @Param courseCode formData string true "Course Code" example(CENG301)
// @Param title formData string true "Title" example("Midterm 1 - Updated")
// @Param content formData string true "Content" example("Updated exam content...")
// @Param files formData file false "Exam files (PDF, image, etc.)" collectionFormat(multi)
// @Param removeFileIds formData string false "Comma-separated list of file IDs to remove" example("1,2,3")
// @Success 200 {object} dto.APIResponse{data=dto.PastExamResponse} "Past exam successfully updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format or validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User does not have permission to update this exam"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /pastexams/{id} [put]
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
	userID, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Invalid userID type")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Convert request to model
	pastExam := &models.PastExam{
		ID:           id,
		Year:         req.Year,
		Term:         models.Term(req.Term),
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
	}

	// Call service to update past exam
	err = c.pastExamService.UpdatePastExam(ctx, pastExam, userIDInt, nil)
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// Process file removals
	if removeFileIdsStr := ctx.PostForm("removeFileIds"); removeFileIdsStr != "" {
		fileIdStrs := strings.Split(removeFileIdsStr, ",")
		for _, fileIdStr := range fileIdStrs {
			fileId, err := strconv.ParseInt(strings.TrimSpace(fileIdStr), 10, 64)
			if err != nil {
				logger.Warn().Err(err).Str("fileIdStr", fileIdStr).Msg("Invalid file ID for removal")
				continue
			}

			err = c.pastExamService.RemoveFileFromPastExam(ctx, id, fileId, userIDInt)
			if err != nil {
				logger.Error().Err(err).Int64("examId", id).Int64("fileId", fileId).Msg("Error removing file from past exam")
				// Continue with other files if one removal fails
			}
		}
	}

	// Handle multiple file uploads
	form, err := ctx.MultipartForm()
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		logger.Error().Err(err).Msg("Error retrieving multipart form")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Error processing file upload")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// If we have files, process them
	uploadedFiles := []*models.File{}
	if form != nil && form.File != nil {
		files := form.File["files"]
		for _, fileHeader := range files {
			// Save the file
			savedFilePath, err := c.fileStorage.SaveFile(fileHeader)
			if err != nil {
				logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error saving uploaded file")
				// Continue with next file if one fails
				continue
			}

			// Create file record in database
			savedFile := &models.File{
				FileName:     fileHeader.Filename,
				FilePath:     savedFilePath,
				FileURL:      c.fileStorage.GetFileURL(savedFilePath),
				FileSize:     fileHeader.Size,
				FileType:     fileHeader.Header.Get("Content-Type"),
				ResourceType: models.FileTypePastExam,
				ResourceID:   id,
				UploadedBy:   userIDInt,
			}

			// Add file to uploaded files collection
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	// If we have any files, associate them with the past exam
	for _, file := range uploadedFiles {
		fileID, err := c.pastExamService.AddFileToPastExam(ctx, id, file)
		if err != nil {
			logger.Error().Err(err).Int64("examId", id).Str("filename", file.FileName).Msg("Error attaching file to past exam")
			// Continue with next file if one fails
		}
		file.ID = fileID
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
// @Summary Delete a past exam (Instructor only, Owner only)
// @Description Delete a past exam by its ID. Requires instructor role and ownership.
// @Tags pastexams
// @Produce json
// @Security BearerAuth
// @Param id path int true "Past Exam ID" Format(int64) example(1)
// @Success 200 {object} dto.APIResponse "Past exam successfully deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid Past Exam ID format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - User is not an instructor or not the owner"
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

	// First get the exam to retrieve the file path before deletion
	pastExam, err := c.pastExamService.GetPastExamByID(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrPastExamNotFound) {
			// If exam doesn't exist, just return the error
			handlePastExamError(ctx, err)
			return
		}
		// For other errors, log and continue with deletion attempt
		logger.Warn().Err(err).Int64("examID", id).Msg("Error getting exam details before deletion")
	}

	// Call service to delete past exam from database
	err = c.pastExamService.DeletePastExam(ctx, id, userID.(int64))
	if err != nil {
		handlePastExamError(ctx, err)
		return
	}

	// If we have the exam details and it had a file, delete the file
	if pastExam != nil && pastExam.Files != nil && len(pastExam.Files) > 0 {
		for _, file := range pastExam.Files {
			fileErr := c.fileStorage.DeleteFile(file.FilePath)
			if fileErr != nil {
				// Log the error but don't fail the operation since the database record is already deleted
				logger.Error().Err(fileErr).Str("filePath", file.FilePath).Msg("Failed to delete file from filesystem")
			} else {
				logger.Info().Str("filePath", file.FilePath).Msg("Successfully deleted file from filesystem")
			}
		}
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success:   true,
		Message:   "Past exam deleted successfully",
		Data:      nil,
		Timestamp: time.Now(),
	})
}
