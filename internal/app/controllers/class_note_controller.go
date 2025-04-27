package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	// "time" // Removed, timestamp handled by common DTO response

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/auth" // Still needed for error checking
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto" // Added DTO import

	// Needed for mapping pagination

	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/filestorage" // Import filestorage
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// ClassNoteController handles API requests related to class notes.
type ClassNoteController struct {
	classNoteService services.ClassNoteService
	fileStorage      *filestorage.LocalStorage // Add file storage service
}

// NewClassNoteController creates a new ClassNoteController.
func NewClassNoteController(service services.ClassNoteService, fileStorage *filestorage.LocalStorage) *ClassNoteController {
	return &ClassNoteController{
		classNoteService: service,
		fileStorage:      fileStorage,
	}
}

// --- Response Wrapper --- // Removed local wrappers, will use dto.APIResponse, dto.ErrorDetail etc.
/*
// Standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type APIError struct {
	Code    int    `json:"code,omitempty"` // Optional: Internal error code
	Message string `json:"message"`
}

func NewSuccessResponse(data interface{}, message string) APIResponse {
	if message == "" {
		message = "Operation completed successfully"
	}
	return APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func NewErrorResponse(statusCode int, err error, message string) APIResponse {
	if message == "" {
		message = "An error occurred"
	}
	// Use default error message if err is nil, otherwise use err.Error()
	errMsg := "Unknown error" // Default
	if err != nil {
		errMsg = err.Error()
	}
	apiErr := &APIError{Message: errMsg}

	// Map specific service/auth errors to messages/status codes
	switch {
	case errors.Is(err, services.ErrClassNotFound), errors.Is(err, auth.ErrResourceNotFound): // Check service & auth not found
		statusCode = http.StatusNotFound
		message = "Resource not found"
		apiErr.Message = errMsg
	case errors.Is(err, services.ErrNoteDepartmentNotFound):
		statusCode = http.StatusBadRequest // Department not found implies bad request data
		message = "Department not found"
		apiErr.Message = errMsg
	case errors.Is(err, auth.ErrPermissionDenied): // Check auth permission denied
		statusCode = http.StatusForbidden
		message = "Permission denied"
		apiErr.Message = errMsg
		// NOTE: middleware errors (401 Unauthorized) are handled by the middleware itself and won't reach here typically.
	}

	// Log internal server errors (anything not handled above and >= 500)
	if statusCode >= 500 {
		logger.Error().Err(err).Msg(message)
		apiErr.Message = "An internal server error occurred." // Don't expose internal details
	}

	return APIResponse{
		Success:   false,
		Message:   message,
		Error:     apiErr,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
*/

// --- Controller Methods ---

// handleClassNoteError maps service errors to HTTP status codes and dto.ErrorDetail
func handleClassNoteError(ctx *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An internal server error occurred")
	errorDetail = errorDetail.WithDetails(err.Error()) // Start with the raw error

	switch {
	case errors.Is(err, services.ErrClassNotFound), errors.Is(err, auth.ErrResourceNotFound):
		statusCode = http.StatusNotFound
		errorDetail = dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Class note not found")
		errorDetail = errorDetail.WithDetails("The requested class note does not exist or you don't have access.") // Generic not found/access denied
	case errors.Is(err, services.ErrNoteDepartmentNotFound):
		statusCode = http.StatusBadRequest // Department not found usually means bad input
		errorDetail = dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid Department ID")
		errorDetail = errorDetail.WithDetails("The specified department ID does not exist.")
	case errors.Is(err, auth.ErrPermissionDenied):
		statusCode = http.StatusForbidden
		errorDetail = dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Permission Denied")
		errorDetail = errorDetail.WithDetails("You do not have permission to perform this action.")
	default:
		// Log unexpected internal errors
		logger.Error().Err(err).Msg("Unhandled error in class note controller")
		// Keep generic message for the client
		errorDetail = dto.NewErrorDetail(dto.ErrorCodeInternalServer, "An unexpected error occurred while processing your request")

	}

	ctx.JSON(statusCode, dto.NewErrorResponse(errorDetail)) // Use common error response constructor
}

// mapServiceNoteToDTO converts a single service note response to a DTO response
func mapServiceNoteToDTO(serviceResp *services.ClassNoteResponse) dto.ClassNoteResponse {
	if serviceResp == nil {
		return dto.ClassNoteResponse{}
	}
	return dto.ClassNoteResponse{
		ID:                serviceResp.ID,
		Year:              serviceResp.Year,
		Term:              string(serviceResp.Term),
		FacultyID:         serviceResp.FacultyID,
		FacultyName:       serviceResp.FacultyName,
		DepartmentID:      serviceResp.DepartmentID,
		DepartmentName:    serviceResp.DepartmentName,
		CourseCode:        serviceResp.CourseCode,
		Title:             serviceResp.Title,
		Content:           serviceResp.Content,
		UploaderName:      serviceResp.UploaderName,
		UploaderEmail:     serviceResp.UploaderEmail,
		UploadedByStudent: serviceResp.UploadedByStudent,
		CreatedAt:         serviceResp.CreatedAt,
		UpdatedAt:         serviceResp.UpdatedAt,
	}
}

// mapServiceNotesToDTO converts a slice of service note responses to a slice of DTO responses
func mapServiceNotesToDTO(serviceNotes []*services.ClassNoteResponse) []dto.ClassNoteResponse {
	dtoNotes := make([]dto.ClassNoteResponse, len(serviceNotes))
	for i, serviceNote := range serviceNotes {
		dtoNotes[i] = mapServiceNoteToDTO(serviceNote)
	}
	return dtoNotes
}

// GetAllClassNotes godoc
// @Summary Get all class notes
// @Description Retrieves a list of class notes with optional filtering and pagination.
// @Tags ClassNotes
// @Accept json
// @Produce json
// @Param facultyId query int false "Filter by Faculty ID" example(1)
// @Param departmentId query int false "Filter by Department ID" example(1)
// @Param courseCode query string false "Filter by Course Code" example(CENG304)
// @Param year query int false "Filter by Year" example(2024)
// @Param term query string false "Filter by Term (FALL or SPRING)" example(SPRING)
// @Param sortBy query string false "Sort field (createdAt, year, term, courseCode, title)" default(createdAt)
// @Param sortOrder query string false "Sort order (ASC or DESC)" default(DESC)
// @Param page query int false "Page number (0-based)" default(0) minimum(0)
// @Param size query int false "Page size" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteListResponse} "Class notes retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request (e.g., invalid query parameters)"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /class-notes [get]
func (ctrl *ClassNoteController) GetAllClassNotes(c *gin.Context) {
	var serviceParams services.GetAllNotesRequest // Service still expects its own request struct

	// --- Parameter Parsing (Mostly unchanged) ---
	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil || page < 0 {
		// Handle invalid page param - return Bad Request
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid page parameter")
		errorDetail = errorDetail.WithDetails("Page must be a non-negative integer.")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}
	serviceParams.Page = page // Pass 0-based page directly

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size <= 0 || size > 100 { // Add max size check
		// Handle invalid size param - return Bad Request
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid size parameter")
		errorDetail = errorDetail.WithDetails("Size must be a positive integer between 1 and 100.")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}
	serviceParams.Size = size

	if facultyIDStr := c.Query("facultyId"); facultyIDStr != "" {
		if id, err := strconv.ParseInt(facultyIDStr, 10, 64); err == nil && id > 0 {
			serviceParams.FacultyID = &id
		} else {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid facultyId parameter")
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			return
		}
	}
	if departmentIDStr := c.Query("departmentId"); departmentIDStr != "" {
		if id, err := strconv.ParseInt(departmentIDStr, 10, 64); err == nil && id > 0 {
			serviceParams.DepartmentID = &id
		} else {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid departmentId parameter")
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			return
		}
	}
	if courseCode := c.Query("courseCode"); courseCode != "" {
		// Optional: Add validation for course code format if needed
		sanitizedCode := strings.ToUpper(strings.TrimSpace(courseCode))
		serviceParams.CourseCode = &sanitizedCode
	}
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil && year > 1900 && year < 2100 { // Example validation
			serviceParams.Year = &year
		} else {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid year parameter")
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			return
		}
	}
	if term := c.Query("term"); term != "" {
		upperTerm := strings.ToUpper(term)
		if upperTerm == string(models.TermFall) || upperTerm == string(models.TermSpring) {
			serviceParams.Term = &upperTerm // Service expects string, validation done there
		} else {
			errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid term parameter")
			errorDetail = errorDetail.WithDetails("Term must be FALL or SPRING")
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
			return
		}
	}

	// Basic validation for sortBy and sortOrder
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	allowedSortBy := map[string]bool{"createdAt": true, "year": true, "term": true, "courseCode": true, "title": true}
	if !allowedSortBy[sortBy] {
		sortBy = "createdAt" // Default to safe value if invalid
	}
	serviceParams.SortBy = sortBy

	sortOrder := strings.ToUpper(c.DefaultQuery("sortOrder", "DESC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC" // Default to safe value
	}
	serviceParams.SortOrder = sortOrder
	// --- End Parameter Parsing ---

	// Call service
	serviceResponse, err := ctrl.classNoteService.GetAllClassNotes(c.Request.Context(), &serviceParams)
	if err != nil {
		handleClassNoteError(c, err) // Use the new error handler
		return
	}

	// Map service response to DTO response
	dtoResponse := dto.ClassNoteListResponse{
		Notes:      mapServiceNotesToDTO(serviceResponse.Notes),
		Pagination: serviceResponse.Pagination, // Assign the value directly from service response
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(dtoResponse, "Class notes retrieved successfully")) // Use common success response
}

// GetClassNoteByID godoc
// @Summary Get class note by ID
// @Description Retrieves a specific class note by its ID.
// @Tags ClassNotes
// @Accept json
// @Produce json
// @Param noteId path int true "Class Note ID" Format(int64) example(15)
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteResponse} "Class note retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request (e.g., invalid ID)"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /class-notes/{noteId} [get]
func (ctrl *ClassNoteController) GetClassNoteByID(c *gin.Context) {
	noteIDStr := c.Param("noteId")
	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil || noteID <= 0 {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid Class Note ID")
		errorDetail = errorDetail.WithDetails("Note ID must be a positive integer.")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service
	serviceResponse, err := ctrl.classNoteService.GetClassNoteByID(c.Request.Context(), noteID)
	if err != nil {
		handleClassNoteError(c, err) // Use the new error handler
		return
	}

	// Map service response to DTO response
	dtoResponse := mapServiceNoteToDTO(serviceResponse)

	c.JSON(http.StatusOK, dto.NewSuccessResponse(dtoResponse, "Class note retrieved successfully")) // Use common success response
}

// CreateClassNote godoc
// @Summary Create a new class note
// @Description Create a new class note with the optional file upload
// @Tags ClassNotes
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param year formData int true "Year" example(2024)
// @Param term formData string true "Term (FALL or SPRING)" example(SPRING)
// @Param departmentId formData int true "Department ID" example(1)
// @Param courseCode formData string true "Course Code" example(CENG304)
// @Param title formData string true "Title" example("Lecture Notes - Week 5")
// @Param content formData string true "Content" example("These notes cover...")
// @Param files formData file false "Files to upload (PDFs, images, etc.)" collectionFormat(multi)
// @Success 201 {object} dto.APIResponse{data=dto.ClassNoteResponse} "Class note created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request (e.g., invalid input data)"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized (missing or invalid token)"
// @Failure 403 {object} dto.ErrorResponse "Forbidden (not permitted to create)"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /class-notes [post]
func (ctrl *ClassNoteController) CreateClassNote(c *gin.Context) {
	// Parse form data instead of JSON
	var req dto.CreateClassNoteRequest
	if err := c.ShouldBind(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid form data")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get userID from context
	userID, exists := c.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Invalid userID type")
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Create class note
	classNoteReq := &services.CreateClassNoteRequest{
		Year:         req.Year,
		Term:         req.Term,
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
	}

	// Call service to create class note
	noteResponse, err := ctrl.classNoteService.CreateClassNote(c.Request.Context(), userIDInt, classNoteReq)
	if err != nil {
		handleClassNoteError(c, err)
		return
	}

	// Handle multiple file uploads if any
	form, err := c.MultipartForm()
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		logger.Error().Err(err).Msg("Error retrieving multipart form")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Error processing file upload")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Process files if we have any
	uploadedFiles := []*models.File{}
	if form != nil && form.File != nil {
		files := form.File["files"]
		for _, fileHeader := range files {
			// Save the file
			savedFilePath, err := ctrl.fileStorage.SaveFile(fileHeader)
			if err != nil {
				logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error saving uploaded file")
				// Continue with next file if one fails
				continue
			}

			// Create file record in database
			savedFile := &models.File{
				FileName:     fileHeader.Filename,
				FilePath:     savedFilePath,
				FileURL:      ctrl.fileStorage.GetFileURL(savedFilePath),
				FileSize:     fileHeader.Size,
				FileType:     fileHeader.Header.Get("Content-Type"),
				ResourceType: models.FileTypeClassNote,
				ResourceID:   noteResponse.ID,
				UploadedBy:   userIDInt,
			}

			// Add file to uploaded files collection
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	// If we have any files, associate them with the class note
	for _, file := range uploadedFiles {
		fileID, err := ctrl.classNoteService.AddFileToClassNote(c.Request.Context(), noteResponse.ID, file)
		if err != nil {
			logger.Error().Err(err).Int64("noteId", noteResponse.ID).Str("filename", file.FileName).Msg("Error attaching file to class note")
			// Continue with next file if one fails
		}
		file.ID = fileID
	}

	// Get updated class note with all details
	updatedNoteResponse, err := ctrl.classNoteService.GetClassNoteByID(c.Request.Context(), noteResponse.ID)
	if err != nil {
		handleClassNoteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(mapServiceNoteToDTO(updatedNoteResponse), "Class note created successfully"))
}

// UpdateClassNote godoc
// @Summary Update an existing class note
// @Description Update a class note that the user owns
// @Tags ClassNotes
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param noteId path int true "Class Note ID" Format(int64) example(15)
// @Param year formData int true "Year" example(2024)
// @Param term formData string true "Term (FALL or SPRING)" example(SPRING)
// @Param departmentId formData int true "Department ID" example(1)
// @Param courseCode formData string true "Course Code" example(CENG304)
// @Param title formData string true "Title" example("Updated Lecture Notes - Week 5")
// @Param content formData string true "Content" example("Updated notes...")
// @Param files formData file false "Files to upload (PDFs, images, etc.)" collectionFormat(multi)
// @Param removeFileIds formData string false "Comma-separated list of file IDs to remove" example("1,2,3")
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteResponse} "Class note updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request (e.g., invalid input data)"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized (missing or invalid token)"
// @Failure 403 {object} dto.ErrorResponse "Forbidden (not the owner of the note)"
// @Failure 404 {object} dto.ErrorResponse "Not Found (class note doesn't exist)"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /class-notes/{noteId} [put]
func (ctrl *ClassNoteController) UpdateClassNote(c *gin.Context) {
	// Get noteId from path parameters
	noteIdStr := c.Param("noteId")
	noteId, err := strconv.ParseInt(noteIdStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid note ID format")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Parse form data instead of JSON
	var req dto.UpdateClassNoteRequest
	if err := c.ShouldBind(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid form data")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get userID from context
	userID, exists := c.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userID.(int64)
	if !ok {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "Invalid userID type")
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Update class note
	updateNoteReq := &services.UpdateClassNoteRequest{
		Year:         req.Year,
		Term:         req.Term,
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
	}

	// Call service to update class note
	_, err = ctrl.classNoteService.UpdateClassNote(c.Request.Context(), userIDInt, noteId, updateNoteReq)
	if err != nil {
		handleClassNoteError(c, err)
		return
	}

	// Process file removals if any
	if removeFileIdsStr := c.PostForm("removeFileIds"); removeFileIdsStr != "" {
		fileIdStrs := strings.Split(removeFileIdsStr, ",")
		for _, fileIdStr := range fileIdStrs {
			fileId, err := strconv.ParseInt(strings.TrimSpace(fileIdStr), 10, 64)
			if err != nil {
				logger.Warn().Err(err).Str("fileIdStr", fileIdStr).Msg("Invalid file ID for removal")
				continue
			}

			err = ctrl.classNoteService.RemoveFileFromClassNote(c.Request.Context(), noteId, fileId, userIDInt)
			if err != nil {
				logger.Error().Err(err).Int64("noteId", noteId).Int64("fileId", fileId).Msg("Error removing file from class note")
				// Continue with other files if one removal fails
			}
		}
	}

	// Handle multiple file uploads if any
	form, err := c.MultipartForm()
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		logger.Error().Err(err).Msg("Error retrieving multipart form")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Error processing file upload")
		errorDetail = errorDetail.WithDetails(err.Error())
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Process files if we have any
	uploadedFiles := []*models.File{}
	if form != nil && form.File != nil {
		files := form.File["files"]
		for _, fileHeader := range files {
			// Save the file
			savedFilePath, err := ctrl.fileStorage.SaveFile(fileHeader)
			if err != nil {
				logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Error saving uploaded file")
				// Continue with next file if one fails
				continue
			}

			// Create file record in database
			savedFile := &models.File{
				FileName:     fileHeader.Filename,
				FilePath:     savedFilePath,
				FileURL:      ctrl.fileStorage.GetFileURL(savedFilePath),
				FileSize:     fileHeader.Size,
				FileType:     fileHeader.Header.Get("Content-Type"),
				ResourceType: models.FileTypeClassNote,
				ResourceID:   noteId,
				UploadedBy:   userIDInt,
			}

			// Add file to uploaded files collection
			uploadedFiles = append(uploadedFiles, savedFile)
		}
	}

	// If we have any files, associate them with the class note
	for _, file := range uploadedFiles {
		fileID, err := ctrl.classNoteService.AddFileToClassNote(c.Request.Context(), noteId, file)
		if err != nil {
			logger.Error().Err(err).Int64("noteId", noteId).Str("filename", file.FileName).Msg("Error attaching file to class note")
			// Continue with next file if one fails
		}
		file.ID = fileID
	}

	// Get updated class note with all details
	updatedNoteResponse, err := ctrl.classNoteService.GetClassNoteByID(c.Request.Context(), noteId)
	if err != nil {
		handleClassNoteError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(mapServiceNoteToDTO(updatedNoteResponse), "Class note updated successfully"))
}

// DeleteClassNote godoc
// @Summary Delete a class note
// @Description Deletes an existing class note. Requires authentication and ownership.
// @Tags ClassNotes
// @Accept json
// @Produce json
// @Param noteId path int true "Class Note ID to delete" Format(int64) example(15)
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "Class note deleted successfully" // Changed data type
// @Failure 400 {object} dto.ErrorResponse "Bad Request (e.g., invalid ID)"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized (handled by middleware)" // Middleware should handle this
// @Failure 403 {object} dto.ErrorResponse "Forbidden (not owner)"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /class-notes/{noteId} [delete]
func (ctrl *ClassNoteController) DeleteClassNote(c *gin.Context) {
	// Get note ID from path
	noteIDStr := c.Param("noteId")
	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil || noteID <= 0 {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid Class Note ID")
		errorDetail = errorDetail.WithDetails("Note ID must be a positive integer.")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get userID from context
	userIDAny, ok := c.Get("userID")
	if !ok {
		err := errors.New("user ID not found in context after auth middleware")
		logger.Error().Err(err).Msg("Missing userID in context")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Authentication context error")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userIDAny.(int64)
	if !ok {
		err := errors.New("invalid user ID type in context")
		logger.Error().Err(err).Interface("userIDValue", userIDAny).Msg("Invalid userID type in context")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid authentication context")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to delete the note from database
	err = ctrl.classNoteService.DeleteClassNote(c.Request.Context(), userIDInt, noteID)
	if err != nil {
		handleClassNoteError(c, err) // Use the new error handler
		return
	}

	// Use common success response with a simple message DTO
	c.JSON(http.StatusOK, dto.NewSuccessResponse(dto.SuccessResponse{Message: "Note deleted successfully"}, "Class note deleted successfully"))
}

// GetMyClassNotes godoc
// @Summary Get my class notes
// @Description Retrieves all class notes uploaded by the currently authenticated user.
// @Tags ClassNotes
// @Accept json
// @Produce json
// @Success 200 {object} dto.APIResponse{data=[]dto.ClassNoteResponse} "Your class notes retrieved successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized (handled by middleware)" // Middleware should handle this
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /class-notes/my-notes [get]
func (ctrl *ClassNoteController) GetMyClassNotes(c *gin.Context) {
	// Get userID from context
	userIDAny, ok := c.Get("userID")
	if !ok {
		err := errors.New("user ID not found in context after auth middleware")
		logger.Error().Err(err).Msg("Missing userID in context")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Authentication context error")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}
	userIDInt, ok := userIDAny.(int64)
	if !ok {
		err := errors.New("invalid user ID type in context")
		logger.Error().Err(err).Interface("userIDValue", userIDAny).Msg("Invalid userID type in context")
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid authentication context")
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service
	serviceResponse, err := ctrl.classNoteService.GetMyClassNotes(c.Request.Context(), userIDInt)
	if err != nil {
		handleClassNoteError(c, err) // Use the new error handler
		return
	}

	// Map service response to DTO response
	dtoResponse := mapServiceNotesToDTO(serviceResponse) // Use the slice mapping helper

	c.JSON(http.StatusOK, dto.NewSuccessResponse(dtoResponse, "Your class notes retrieved successfully")) // Use common success response
}
