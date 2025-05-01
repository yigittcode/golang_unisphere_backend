package controllers

import (
	"net/http"
	"strconv"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
)

// parseIDParam parses an ID parameter from the request path
func parseIDParam(ctx *gin.Context, paramName string) (int64, error) {
	idStr := ctx.Param(paramName)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

// ClassNoteController handles class note operations
type ClassNoteController struct {
	classNoteService services.ClassNoteService
	fileStorage      *filestorage.LocalStorage
}

// NewClassNoteController creates a new ClassNoteController
func NewClassNoteController(classNoteService services.ClassNoteService, fileStorage *filestorage.LocalStorage) *ClassNoteController {
	return &ClassNoteController{
		classNoteService: classNoteService,
		fileStorage:      fileStorage,
	}
}

// GetAllNotes godoc
// @Summary Get all class notes
// @Description Get a list of all class notes with optional filtering
// @Tags class-notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param departmentId query int false "Filter by department ID"
// @Param courseCode query string false "Filter by course code"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteListResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes [get]
func (c *ClassNoteController) GetAllNotes(ctx *gin.Context) {
	// Parse filter parameters
	var filter dto.ClassNoteFilterRequest
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid filter parameters")))
		return
	}

	// Get notes with pagination
	notes, err := c.classNoteService.GetAllNotes(ctx, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get class notes")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(notes))
}

// GetNoteByID godoc
// @Summary Get a class note by ID
// @Description Get detailed information about a specific class note
// @Tags class-notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Class note ID"
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes/{noteId} [get]
func (c *ClassNoteController) GetNoteByID(ctx *gin.Context) {
	// Parse ID from path
	id, err := parseIDParam(ctx, "noteId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid class note ID")))
		return
	}

	// Get note
	note, err := c.classNoteService.GetNoteByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get class note")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(note))
}

// CreateNote godoc
// @Summary Create a new class note
// @Description Create a new class note with file upload
// @Tags class-notes
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param courseCode formData string true "Course code"
// @Param title formData string true "Title"
// @Param description formData string true "Description"
// @Param departmentId formData int true "Department ID"
// @Param files formData file false "Files to upload" collectionFormat multi
// @Success 201 {object} dto.APIResponse{data=dto.ClassNoteResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes [post]
func (c *ClassNoteController) CreateNote(ctx *gin.Context) {
	// Parse request data
	var req dto.CreateClassNoteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format")))
		return
	}

	// Get files (if any)
	form, err := ctx.MultipartForm()
	var files []*multipart.FileHeader

	if err == nil && form != nil && form.File != nil {
		files = form.File["files"]
	}

	// Create note (even without files)
	note, err := c.classNoteService.CreateNote(ctx, &req, files)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to create class note")))
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(note))
}

// UpdateNote godoc
// @Summary Update a class note
// @Description Update an existing class note
// @Tags class-notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Class note ID"
// @Param request body dto.UpdateClassNoteRequest true "Update data"
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes/{noteId} [put]
func (c *ClassNoteController) UpdateNote(ctx *gin.Context) {
	// Parse ID from path
	id, err := parseIDParam(ctx, "noteId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid class note ID")))
		return
	}

	// Parse request data
	var req dto.UpdateClassNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format")))
		return
	}

	// Update note
	note, err := c.classNoteService.UpdateNote(ctx, id, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to update class note")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(note))
}

// DeleteNote godoc
// @Summary Delete a class note
// @Description Delete an existing class note
// @Tags class-notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Class note ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes/{noteId} [delete]
func (c *ClassNoteController) DeleteNote(ctx *gin.Context) {
	// Parse ID from path
	id, err := parseIDParam(ctx, "noteId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid class note ID")))
		return
	}

	// Delete note
	err = c.classNoteService.DeleteNote(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete class note")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "Class note deleted successfully"}))
}

// AddFilesToNote godoc
// @Summary Add files to an existing class note
// @Description Add multiple files to an existing class note
// @Tags class-notes
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Class note ID"
// @Param files formData file true "Files to upload" collectionFormat multi
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes/{noteId}/files [post]
func (c *ClassNoteController) AddFilesToNote(ctx *gin.Context) {
	// Parse ID from path
	id, err := parseIDParam(ctx, "noteId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid class note ID")))
		return
	}

	// Get files
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid form data")))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "No files uploaded")))
		return
	}

	// Add files to note
	updatedNote, err := c.classNoteService.AddFilesToNote(ctx, id, files)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to add files to class note")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(updatedNote))
}
