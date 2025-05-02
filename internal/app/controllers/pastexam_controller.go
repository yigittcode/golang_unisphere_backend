package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/helpers"
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

// GetAllPastExams handles retrieving all past exams with optional filtering
// @Summary Get all past exams
// @Description Retrieves a list of past exams with optional filtering and pagination. Available to all authenticated users.
// @Tags past-exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param facultyId query int false "Filter by faculty ID"
// @Param departmentId query int false "Filter by department ID"
// @Param courseCode query string false "Filter by course code"
// @Param year query int false "Filter by year"
// @Param term query string false "Filter by term (FALL, SPRING)"
// @Param sortBy query string false "Sort field (year, term, courseCode, title, departmentName, facultyName, instructorName, createdAt, updatedAt)"
// @Param sortOrder query string false "Sort order (ASC, DESC)"
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param pageSize query int false "Page size (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.PastExamListResponse} "Past exams retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams [get]
func (c *PastExamController) GetAllPastExams(ctx *gin.Context) {
	// Parse pagination parameters using helper
	page, pageSize := helpers.ParsePaginationParams(ctx)

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
	if facultyID, ok := filters["facultyId"].(int64); ok {
		filter.FacultyID = &facultyID
	}
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
// @Success 200 {object} dto.APIResponse{data=dto.PastExamResponse} "Past exam retrieved successfully"
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

// CreatePastExam godoc
// @Summary Create a new past exam
// @Description Create a new past exam with file upload. Only instructors can create past exams. The current authenticated instructor will be set as the instructor of the exam.
// @Tags past-exams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param year formData int true "Year"
// @Param term formData string true "Term (FALL, SPRING)"
// @Param departmentId formData int true "Department ID"
// @Param courseCode formData string true "Course code"
// @Param title formData string true "Title"
// @Param files formData file false "Exam files (can upload multiple)"
// @Success 201 {object} dto.APIResponse{data=dto.PastExamResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User does not have instructor role"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /past-exams [post]
func (c *PastExamController) CreatePastExam(ctx *gin.Context) {
	fmt.Println("********* CreatePastExam BAŞLANGIÇ *********")
	
	var req dto.CreatePastExamRequest
	if err := ctx.ShouldBind(&req); err != nil {
		fmt.Printf("Error binding request: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format").WithDetails(err.Error())))
		return
	}

	// Validate term
	termValue := dto.Term(req.Term)
	if termValue != dto.TermFall && termValue != dto.TermSpring {
		fmt.Printf("Invalid term value: %s\n", req.Term)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid term value. Must be FALL or SPRING")))
		return
	}

	// Get files (optional)
	var files []*multipart.FileHeader
	
	// Get files from the multipart form
	form, err := ctx.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		if uploadedFiles, ok := form.File["files"]; ok && len(uploadedFiles) > 0 {
			files = uploadedFiles
			fmt.Printf("Files included: %d files\n", len(files))
		} else {
			fmt.Println("No files provided in the 'files' field")
		}
	} else {
		fmt.Println("No files provided or error getting files, continuing without files")
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		fmt.Println("User ID not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")))
		return
	}
	
	// Get user role
	roleType, exists := ctx.Get("roleType") 
	if !exists {
		fmt.Println("User role not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User role not found")))
		return
	}
	
	// Convert to string and check role
	roleStr, ok := roleType.(string)
	if !ok || roleStr != string(models.RoleInstructor) {
		fmt.Printf("User has invalid role for creating past exam: %v\n", roleStr)
		ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeForbidden, "Only instructors can create past exams")))
		return
	}
	
	fmt.Printf("Creating past exam with instructor ID: %v\n", userID)
	fmt.Printf("Request data: %+v\n", req)

	// Create exam
	exam, err := c.pastExamService.CreateExam(ctx, &req, files)
	if err != nil {
		fmt.Printf("Error creating past exam: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to create past exam").WithDetails(err.Error())))
		return
	}

	fmt.Println("********* CreatePastExam BAŞARILI *********")
	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(exam))
}

// UpdatePastExam handles updating an existing past exam
// @Summary Update a past exam
// @Description Updates an existing past exam with the provided information. Only instructors can update past exams. The instructor must be the owner of the exam.
// @Tags past-exams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Past exam ID"
// @Param year formData int false "Year of the exam"
// @Param term formData string false "Term of the exam (FALL, SPRING)"
// @Param departmentId formData int false "Department ID"
// @Param courseCode formData string false "Course code"
// @Param title formData string false "Exam title"
//
// @Param file formData file false "Exam file"
// @Success 200 {object} dto.APIResponse{data=dto.PastExamResponse} "Past exam updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.ErrorResponse "Forbidden: User does not have instructor role or is not the creator"
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
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "Past exam deleted successfully" 
// @Failure 400 {object} dto.ErrorResponse "Invalid past exam ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.ErrorResponse "Forbidden: User does not have instructor role or is not the creator"
// @Failure 404 {object} dto.ErrorResponse "Past exam not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /past-exams/{id} [delete]
func (c *PastExamController) DeletePastExam(ctx *gin.Context) {
	fmt.Println("********* DeletePastExam BAŞLANGIÇ *********")
	
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid past exam ID")
		errorDetail = errorDetail.WithDetails("Past exam ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	fmt.Printf("Attempting to delete past exam with ID: %d\n", id)

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
		fmt.Printf("Error deleting past exam: %v\n", err)
		
		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrPastExamNotFound):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Past exam not found")))
			return
		case errors.Is(err, apperrors.ErrPermissionDenied) || strings.Contains(err.Error(), "unauthorized"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "You don't have permission to delete this past exam")))
			return
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete past exam").WithDetails(err.Error())))
			return
		}
	}

	// Delete associated files - now handled by the service
	fmt.Println("********* DeletePastExam BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "Past exam deleted successfully"}))
}

// AddFileToPastExam godoc
// @Summary Add files to an existing past exam
// @Description Add one or more files to an existing past exam
// @Tags past-exams
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Past exam ID"
// @Param files formData file true "Files to upload (can be multiple)"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User does not have instructor role or is not the creator"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /past-exams/{id}/files [post]
func (c *PastExamController) AddFileToPastExam(ctx *gin.Context) {
	fmt.Println("********* AddFileToPastExam BAŞLANGIÇ *********")
	
	// Parse exam ID from path
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid past exam ID")))
		return
	}
	
	fmt.Printf("Adding files to past exam with ID: %d\n", id)

	// Get files
	var files []*multipart.FileHeader
	
	// Get files from the multipart form
	form, err := ctx.MultipartForm()
	if err == nil && form != nil && form.File != nil {
		if uploadedFiles, ok := form.File["files"]; ok && len(uploadedFiles) > 0 {
			files = uploadedFiles
			fmt.Printf("Files included: %d files\n", len(files))
		} else {
			fmt.Println("No files provided in the 'files' field")
		}
	} else {
		fmt.Println("No files provided or error getting files")
	}
	
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "No valid files provided")))
		return
	}

	// Process each file
	successCount := 0
	var lastError error
	
	for _, fileHeader := range files {
		// Add file to past exam
		err = c.pastExamService.AddFileToPastExam(ctx, id, fileHeader)
		if err != nil {
			fmt.Printf("Error adding file '%s' to past exam: %v\n", fileHeader.Filename, err)
			lastError = err
		} else {
			successCount++
		}
	}
	
	if successCount == 0 && lastError != nil {
		// All files failed to upload
		switch {
		case errors.Is(lastError, apperrors.ErrPastExamNotFound):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Past exam not found")))
		case errors.Is(lastError, apperrors.ErrPermissionDenied) || strings.Contains(lastError.Error(), "unauthorized"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "You don't have permission to update this past exam")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to add files to past exam").WithDetails(lastError.Error())))
		}
		return
	}

	fmt.Printf("********* AddFileToPastExam BAŞARILI: %d/%d files added *********\n", successCount, len(files))
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: fmt.Sprintf("%d files added to past exam successfully", successCount)}))
}

// DeleteFileFromPastExam godoc
// @Summary Delete a file from a past exam
// @Description Remove a file from a past exam
// @Tags past-exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Past exam ID"
// @Param fileId path int true "File ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User does not have instructor role or is not the creator"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /past-exams/{id}/files/{fileId} [delete]
func (c *PastExamController) DeleteFileFromPastExam(ctx *gin.Context) {
	fmt.Println("********* DeleteFileFromPastExam BAŞLANGIÇ *********")
	
	// Parse exam ID from path
	examIDStr := ctx.Param("id")
	examID, err := strconv.ParseInt(examIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid past exam ID")))
		return
	}

	// Parse file ID from path
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid file ID")))
		return
	}
	
	fmt.Printf("Deleting file %d from past exam with ID: %d\n", fileID, examID)

	// Delete file from past exam
	err = c.pastExamService.RemoveFileFromPastExam(ctx, examID, fileID)
	if err != nil {
		fmt.Printf("Error deleting file from past exam: %v\n", err)
		
		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrPastExamNotFound):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Past exam not found")))
		case errors.Is(err, apperrors.ErrPermissionDenied) || strings.Contains(err.Error(), "unauthorized"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "You don't have permission to modify this past exam")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete file from past exam").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* DeleteFileFromPastExam BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "File deleted from past exam successfully"}))
}
