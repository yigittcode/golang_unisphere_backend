package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
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
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param pageSize query int false "Page size (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteListResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes [get]
func (c *ClassNoteController) GetAllNotes(ctx *gin.Context) {
	fmt.Println("********* GetAllNotes *********")
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
// @Param content formData string true "Content/text of the note"
// @Param departmentId formData int true "Department ID"
// @Param files formData file false "Files to upload" collectionFormat multi
// @Success 201 {object} dto.APIResponse{data=dto.ClassNoteResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /class-notes [post]
func (c *ClassNoteController) CreateNote(ctx *gin.Context) {
	fmt.Println("********* CreateNote BAŞLANGIÇ *********")
	fmt.Printf("Content-Type: %s\n", ctx.GetHeader("Content-Type"))

	// Middleware validasyonu başarıyla geçti mi kontrol et
	validatedObj, exists := ctx.Get("validatedFormData")
	if exists {
		fmt.Println("Middleware validasyonu başarılı!")
		fmt.Printf("Validated object: %+v\n", validatedObj)
		req := validatedObj.(*dto.CreateClassNoteRequest)
		fmt.Printf("courseCode: %s\n", req.CourseCode)
		fmt.Printf("title: %s\n", req.Title)
		fmt.Printf("description: %s\n", req.Description)
		fmt.Printf("content: %s\n", req.Content)
		fmt.Printf("departmentId: %d\n", req.DepartmentID)
	} else {
		// Form değerlerini kontrol et
		fmt.Println("Validasyon middleware'i çalışmadı, ham verileri kontrol ediyorum:")
		courseCode := ctx.PostForm("courseCode")
		title := ctx.PostForm("title")
		description := ctx.PostForm("description")
		content := ctx.PostForm("content")
		departmentId := ctx.PostForm("departmentId")
		fmt.Printf("courseCode: %s\n", courseCode)
		fmt.Printf("title: %s\n", title)
		fmt.Printf("description: %s\n", description)
		fmt.Printf("content: %s\n", content)
		fmt.Printf("departmentId: %s\n", departmentId)
	}

	// Parse request data - middleware bağlamış olsa bile, bu sıfırdan bağlama yapar
	var req dto.CreateClassNoteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		fmt.Printf("BINDING HATASI: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format").WithDetails(err.Error())))
		return
	}

	fmt.Printf("PARSED REQUEST: %+v\n", req)

	// Get files (if any)
	form, err := ctx.MultipartForm()
	var files []*multipart.FileHeader

	if err == nil && form != nil && form.File != nil {
		files = form.File["files"]
		fmt.Printf("Dosya sayısı: %d\n", len(files))
	} else if err != nil {
		fmt.Printf("MultipartForm hatası: %v\n", err)
	}

	// Create note (even without files)
	note, err := c.classNoteService.CreateNote(ctx, &req, files)
	if err != nil {
		fmt.Printf("Service hatası: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to create class note").WithDetails(err.Error())))
		return
	}

	fmt.Println("********* CreateNote BAŞARILI *********")
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

// GetFileDetails godoc
// @Summary Get file details
// @Description Get detailed information about a specific file (works for any file type)
// @Tags files
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param fileId path int true "File ID"
// @Success 200 {object} dto.APIResponse{data=dto.ClassNoteFileResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /files/{fileId} [get]
func (c *ClassNoteController) GetFileDetails(ctx *gin.Context) {
	// Parse ID from path
	id, err := parseIDParam(ctx, "fileId")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid file ID")))
		return
	}

	// Get file details
	fileDetails, err := c.classNoteService.GetFileDetails(ctx, id)
	if err != nil {
		// Handle common error types
		if errors.Is(err, apperrors.ErrResourceNotFound) {
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "File not found")))
			return
		}

		// Handle other errors
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get file details")))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(fileDetails))
}
