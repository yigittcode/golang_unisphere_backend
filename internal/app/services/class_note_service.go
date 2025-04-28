package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// ClassNoteService defines the interface for class note operations.
type ClassNoteService interface {
	GetAllClassNotes(ctx context.Context, params *GetAllNotesRequest) (*GetAllNotesResponse, error)
	GetClassNoteByID(ctx context.Context, noteID int64) (*ClassNoteResponse, error)
	CreateClassNote(ctx context.Context, userID int64, req *CreateClassNoteRequest) (*ClassNoteResponse, error)
	UpdateClassNote(ctx context.Context, userID int64, noteID int64, req *UpdateClassNoteRequest) (*ClassNoteResponse, error)
	DeleteClassNote(ctx context.Context, userID int64, noteID int64) error
	GetMyClassNotes(ctx context.Context, userID int64) ([]*ClassNoteResponse, error)
	AddFileToClassNote(ctx context.Context, noteID int64, file *models.File) (int64, error)
	RemoveFileFromClassNote(ctx context.Context, noteID, fileID, userID int64) error
	GetClassNoteFiles(ctx context.Context, noteID int64) ([]*models.File, error)
}

// classNoteServiceImpl implements ClassNoteService.
type classNoteServiceImpl struct {
	noteRepo    *repositories.ClassNoteRepository
	deptRepo    *repositories.DepartmentRepository // Needed to validate department/faculty
	authService *auth.AuthorizationService         // Changed to pointer based on typical Go patterns
}

// NewClassNoteService creates a new ClassNoteService.
func NewClassNoteService(noteRepo *repositories.ClassNoteRepository, deptRepo *repositories.DepartmentRepository, authService *auth.AuthorizationService) ClassNoteService {
	return &classNoteServiceImpl{
		noteRepo:    noteRepo,
		deptRepo:    deptRepo,
		authService: authService,
	}
}

// DTOs (Data Transfer Objects) for ClassNote API
// Consider moving these to a separate models/dto package if they grow

// ClassNoteResponse represents the data returned for a single class note.
type ClassNoteResponse struct {
	ID                int64       `json:"id"`
	Year              int         `json:"year"`
	Term              models.Term `json:"term"`
	FacultyID         int64       `json:"facultyId"`
	FacultyName       string      `json:"facultyName"`
	DepartmentID      int64       `json:"departmentId"`
	DepartmentName    string      `json:"departmentName"`
	CourseCode        string      `json:"courseCode"`
	Title             string      `json:"title"`
	Content           string      `json:"content"`
	UploaderName      string      `json:"uploaderName"`
	UploaderEmail     string      `json:"uploaderEmail"`
	UploadedByStudent bool        `json:"uploadedByStudent"`
	CreatedAt         string      `json:"createdAt"`
	UpdatedAt         string      `json:"updatedAt"`
}

// GetAllNotesRequest represents the filter and pagination parameters for getting all notes.
type GetAllNotesRequest struct {
	FacultyID    *int64
	DepartmentID *int64
	CourseCode   *string
	Year         *int
	Term         *string // Keep as string for input, convert to models.Term in service
	SortBy       string
	SortOrder    string
	Page         int
	Size         int
}

// GetAllNotesResponse represents the paginated list of class notes.
type GetAllNotesResponse struct {
	Notes      []*ClassNoteResponse `json:"notes"`
	Pagination dto.PaginationInfo   `json:"pagination"`
}

// CreateClassNoteRequest represents the data needed to create a new class note.
type CreateClassNoteRequest struct {
	Year         int    `json:"year" binding:"required,gte=2000"`
	Term         string `json:"term" binding:"required,oneof=FALL SPRING"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
	CourseCode   string `json:"courseCode" binding:"required,alphanum,uppercase,min=3,max=10"`
	Title        string `json:"title" binding:"required,min=5,max=255"`
	Content      string `json:"content" binding:"required,min=10"`
}

// UpdateClassNoteRequest represents the data needed to update a class note.
type UpdateClassNoteRequest struct {
	Year         int    `json:"year" binding:"required,gte=2000"`
	Term         string `json:"term" binding:"required,oneof=FALL SPRING"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
	CourseCode   string `json:"courseCode" binding:"required,alphanum,uppercase,min=3,max=10"`
	Title        string `json:"title" binding:"required,min=5,max=255"`
	Content      string `json:"content" binding:"required,min=10"`
}

// --- Service Implementation ---

func (s *classNoteServiceImpl) GetAllClassNotes(ctx context.Context, params *GetAllNotesRequest) (*GetAllNotesResponse, error) {
	// Validate and convert term to models.Term
	var termModel *models.Term
	if params.Term != nil {
		t := models.Term(*params.Term)
		if t != models.TermFall && t != models.TermSpring {
			return nil, fmt.Errorf("invalid term: %s", *params.Term)
		}
		termModel = &t
	}

	repoParams := repositories.GetAllNotesParams{
		FacultyID:    params.FacultyID,
		DepartmentID: params.DepartmentID,
		CourseCode:   params.CourseCode,
		Year:         params.Year,
		Term:         termModel,
		SortBy:       params.SortBy,
		SortOrder:    params.SortOrder,
		Page:         params.Page,
		Size:         params.Size,
	}

	notesDetails, pagination, err := s.noteRepo.GetAllClassNotes(ctx, repoParams)
	if err != nil {
		logger.Error().Err(err).Msg("Error getting all class notes from repository")
		return nil, fmt.Errorf("failed to retrieve class notes: %w", err)
	}

	// Map repository details to response DTOs
	responseNotes := make([]*ClassNoteResponse, len(notesDetails))
	for i, note := range notesDetails {
		responseNotes[i] = mapNoteDetailsToResponse(note)
	}

	return &GetAllNotesResponse{
		Notes:      responseNotes,
		Pagination: pagination,
	}, nil
}

func (s *classNoteServiceImpl) GetClassNoteByID(ctx context.Context, noteID int64) (*ClassNoteResponse, error) {
	noteDetails, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrClassNoteNotFound
		}
		logger.Error().Err(err).Int64("noteID", noteID).Msg("Error getting class note by ID from repository")
		return nil, fmt.Errorf("failed to retrieve class note: %w", err)
	}

	return mapNoteDetailsToResponse(noteDetails), nil
}

func (s *classNoteServiceImpl) CreateClassNote(ctx context.Context, userID int64, req *CreateClassNoteRequest) (*ClassNoteResponse, error) {
	// Validate Department exists
	dept, err := s.deptRepo.GetByID(ctx, req.DepartmentID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrDepartmentNotFound
		}
		logger.Error().Err(err).Int64("departmentID", req.DepartmentID).Msg("Error checking department existence")
		return nil, fmt.Errorf("failed to validate department: %w", err)
	}
	if dept == nil { // Should be covered by ErrNotFound
		return nil, apperrors.ErrDepartmentNotFound
	}

	note := &models.ClassNote{
		Year:         req.Year,
		Term:         models.Term(req.Term),
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
		UserID:       userID,
	}

	newNoteID, err := s.noteRepo.CreateClassNote(ctx, note)
	if err != nil {
		// TODO: Handle potential constraint errors (e.g., duplicate course code for term?)
		logger.Error().Err(err).Int64("userID", userID).Msg("Error creating class note in repository")
		return nil, fmt.Errorf("failed to create class note: %w", err)
	}

	// Retrieve the full details of the newly created note
	createdNoteDetails, err := s.noteRepo.GetClassNoteByID(ctx, newNoteID)
	if err != nil {
		// Log error but potentially continue, or return partial response? Best to return error.
		logger.Error().Err(err).Int64("newNoteID", newNoteID).Msg("Error retrieving newly created class note details")
		return nil, fmt.Errorf("failed to retrieve created class note details: %w", err)
	}

	return mapNoteDetailsToResponse(createdNoteDetails), nil
}

func (s *classNoteServiceImpl) UpdateClassNote(ctx context.Context, userID int64, noteID int64, req *UpdateClassNoteRequest) (*ClassNoteResponse, error) {
	// 1. Validate Ownership using AuthorizationService
	err := s.authService.ValidateClassNoteOwnership(ctx, noteID, userID)
	if err != nil {
		// Return specific auth errors directly (ErrResourceNotFound, ErrPermissionDenied)
		return nil, err
	}

	// 2. Validate Department exists (if changed, though request makes it required)
	dept, err := s.deptRepo.GetByID(ctx, req.DepartmentID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrDepartmentNotFound
		}
		logger.Error().Err(err).Int64("departmentID", req.DepartmentID).Msg("Error checking department existence for update")
		return nil, fmt.Errorf("failed to validate department for update: %w", err)
	}
	if dept == nil { // Should be covered by ErrNotFound
		return nil, apperrors.ErrDepartmentNotFound
	}

	// 3. Fetch existing note to preserve fields not in request (like UserID)
	// Although not strictly necessary if repo update only sets provided fields,
	// it's safer if the update logic expects a full model.
	// Alternatively, pass individual fields to repo update.
	existingNote, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrClassNoteNotFound
		}
		logger.Error().Err(err).Int64("noteID", noteID).Msg("Failed to fetch existing class note for update")
		return nil, fmt.Errorf("failed to fetch existing note: %w", err)
	}

	// 4. Prepare updated note model
	noteToUpdate := &models.ClassNote{
		ID:           noteID,
		Year:         req.Year,
		Term:         models.Term(req.Term),
		DepartmentID: req.DepartmentID,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Content:      req.Content,
		UserID:       existingNote.UserID,
		CreatedAt:    existingNote.CreatedAt,
	}

	// 5. Perform update in repository
	err = s.noteRepo.UpdateClassNote(ctx, noteToUpdate)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// This means the note disappeared between ownership check and update, or ID was wrong
			return nil, apperrors.ErrClassNoteNotFound
		}
		logger.Error().Err(err).Int64("noteID", noteID).Msg("Error updating class note in repository")
		return nil, fmt.Errorf("failed to update class note: %w", err)
	}

	// 6. Retrieve updated details
	updatedNoteDetails, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		logger.Error().Err(err).Int64("noteID", noteID).Msg("Error retrieving updated class note details")
		return nil, fmt.Errorf("failed to retrieve updated class note details: %w", err)
	}

	return mapNoteDetailsToResponse(updatedNoteDetails), nil
}

func (s *classNoteServiceImpl) DeleteClassNote(ctx context.Context, userID int64, noteID int64) error {
	// 1. Validate Ownership using AuthorizationService
	err := s.authService.ValidateClassNoteOwnership(ctx, noteID, userID)
	if err != nil {
		return err // Return specific auth errors directly
	}

	// 2. Perform delete in repository
	err = s.noteRepo.DeleteClassNote(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// Note already deleted or never existed (after ownership check? Unlikely but possible)
			return apperrors.ErrClassNoteNotFound
		}
		logger.Error().Err(err).Int64("noteID", noteID).Msg("Error deleting class note from repository")
		return fmt.Errorf("failed to delete class note: %w", err)
	}

	return nil // Success
}

func (s *classNoteServiceImpl) GetMyClassNotes(ctx context.Context, userID int64) ([]*ClassNoteResponse, error) {
	notesDetails, err := s.noteRepo.GetClassNotesByUploaderID(ctx, userID)
	if err != nil {
		// No specific error for empty list, just return empty slice
		logger.Error().Err(err).Int64("userID", userID).Msg("Error getting class notes by uploader ID from repository")
		return nil, fmt.Errorf("failed to retrieve your class notes: %w", err)
	}

	// Map repository details to response DTOs
	responseNotes := make([]*ClassNoteResponse, len(notesDetails))
	for i, note := range notesDetails {
		responseNotes[i] = mapNoteDetailsToResponse(note)
	}

	return responseNotes, nil
}

// AddFileToClassNote adds a file to a class note
func (s *classNoteServiceImpl) AddFileToClassNote(ctx context.Context, noteID int64, file *models.File) (int64, error) {
	// Check if the class note exists
	_, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return 0, apperrors.ErrClassNoteNotFound
		}
		return 0, fmt.Errorf("error checking class note existence: %w", err)
	}

	// Check if user has permission to add files to this class note
	err = s.authService.ValidateClassNoteOwnership(ctx, noteID, file.UploadedBy)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return 0, auth.ErrPermissionDenied
		}
		return 0, fmt.Errorf("authorization validation error: %w", err)
	}

	// Save file metadata in database
	fileRepo := s.noteRepo.GetFileRepository()
	fileID, err := fileRepo.CreateFile(ctx, file)
	if err != nil {
		return 0, fmt.Errorf("error creating file record: %w", err)
	}

	// Associate file with class note
	err = fileRepo.AddFileToClassNote(ctx, noteID, fileID)
	if err != nil {
		return 0, fmt.Errorf("error associating file with class note: %w", err)
	}

	return fileID, nil
}

// RemoveFileFromClassNote removes a file from a class note
func (s *classNoteServiceImpl) RemoveFileFromClassNote(ctx context.Context, noteID, fileID, userID int64) error {
	// Check if the class note exists
	_, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.ErrClassNoteNotFound
		}
		return fmt.Errorf("error checking class note existence: %w", err)
	}

	// Check if user has permission to remove files from this class note
	err = s.authService.ValidateClassNoteOwnership(ctx, noteID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return auth.ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	// Get file repository
	fileRepo := s.noteRepo.GetFileRepository()

	// Remove association between file and class note
	err = fileRepo.RemoveFileFromClassNote(ctx, noteID, fileID)
	if err != nil {
		return fmt.Errorf("error removing file association: %w", err)
	}

	// Optionally delete the file itself if it's not used elsewhere
	// This could be implemented separately if needed

	return nil
}

// GetClassNoteFiles gets all files associated with a class note
func (s *classNoteServiceImpl) GetClassNoteFiles(ctx context.Context, noteID int64) ([]*models.File, error) {
	// Check if the class note exists
	_, err := s.noteRepo.GetClassNoteByID(ctx, noteID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrClassNoteNotFound
		}
		return nil, fmt.Errorf("error checking class note existence: %w", err)
	}

	// Get file repository
	fileRepo := s.noteRepo.GetFileRepository()

	// Get all files associated with this class note
	files, err := fileRepo.GetFilesByClassNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving files: %w", err)
	}

	return files, nil
}

// --- Helper Functions ---

// mapNoteDetailsToResponse converts repository details to API response DTO.
func mapNoteDetailsToResponse(details *repositories.ClassNoteDetails) *ClassNoteResponse {
	if details == nil {
		return nil
	}

	uploaderName := fmt.Sprintf("%s %s", details.UploaderFirstName, details.UploaderLastName)

	resp := &ClassNoteResponse{
		ID:                details.ID,
		Year:              details.Year,
		Term:              details.Term,
		FacultyID:         details.FacultyID,
		FacultyName:       details.FacultyName,
		DepartmentID:      details.DepartmentID,
		DepartmentName:    details.DepartmentName,
		CourseCode:        details.CourseCode,
		Title:             details.Title,
		Content:           details.Content,
		UploaderName:      uploaderName,
		UploaderEmail:     details.UploaderEmail,
		UploadedByStudent: details.UploadedByStudent,
		CreatedAt:         details.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         details.UpdatedAt.Format(time.RFC3339),
	}

	return resp
}

// --- Service now uses apperrors instead of local error definitions ---
// These imports are now centralized in the apperrors package
