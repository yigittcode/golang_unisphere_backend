package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// ClassNoteService defines the interface for class note operations
type ClassNoteService interface {
	GetAllNotes(ctx context.Context, filter *dto.ClassNoteFilterRequest) (*dto.ClassNoteListResponse, error)
	GetNoteByID(ctx context.Context, id int64) (*dto.ClassNoteResponse, error)
	CreateNote(ctx context.Context, req *dto.CreateClassNoteRequest, file *multipart.FileHeader) (*dto.ClassNoteResponse, error)
	UpdateNote(ctx context.Context, id int64, req *dto.UpdateClassNoteRequest) (*dto.ClassNoteResponse, error)
	DeleteNote(ctx context.Context, id int64) error
}

// classNoteServiceImpl implements ClassNoteService
type classNoteServiceImpl struct {
	classNoteRepo  *repositories.ClassNoteRepository
	departmentRepo *repositories.DepartmentRepository
	authzService   *auth.AuthorizationService
}

// NewClassNoteService creates a new ClassNoteService
func NewClassNoteService(
	classNoteRepo *repositories.ClassNoteRepository,
	departmentRepo *repositories.DepartmentRepository,
	authzService *auth.AuthorizationService,
) ClassNoteService {
	return &classNoteServiceImpl{
		classNoteRepo:  classNoteRepo,
		departmentRepo: departmentRepo,
		authzService:   authzService,
	}
}

// GetAllNotes retrieves all class notes with filtering and pagination
func (s *classNoteServiceImpl) GetAllNotes(ctx context.Context, filter *dto.ClassNoteFilterRequest) (*dto.ClassNoteListResponse, error) {
	// Get notes from repository
	notes, total, err := s.classNoteRepo.GetAll(ctx, filter.DepartmentID, filter.CourseCode, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("error getting class notes: %w", err)
	}

	// Convert to response DTOs
	var noteResponses []dto.ClassNoteResponse
	for _, note := range notes {
		noteResponses = append(noteResponses, dto.ClassNoteResponse{
			ID:           note.ID,
			CourseCode:   note.CourseCode,
			Title:        note.Title,
			Description:  note.Description,
			FileID:       note.FileID,
			DepartmentID: note.DepartmentID,
			InstructorID: note.InstructorID,
		})
	}

	// Create response with pagination
	totalPages := (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	return &dto.ClassNoteListResponse{
		ClassNotes: noteResponses,
		PaginationInfo: dto.PaginationInfo{
			CurrentPage: filter.Page,
			PageSize:    filter.PageSize,
			TotalItems:  total,
			TotalPages:  int(totalPages),
		},
	}, nil
}

// GetNoteByID retrieves a class note by ID
func (s *classNoteServiceImpl) GetNoteByID(ctx context.Context, id int64) (*dto.ClassNoteResponse, error) {
	// Get note from repository
	note, err := s.classNoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting class note: %w", err)
	}
	if note == nil {
		return nil, apperrors.ErrClassNoteNotFound
	}

	// Convert to response DTO
	return &dto.ClassNoteResponse{
		ID:           note.ID,
		CourseCode:   note.CourseCode,
		Title:        note.Title,
		Description:  note.Description,
		FileID:       note.FileID,
		DepartmentID: note.DepartmentID,
		InstructorID: note.InstructorID,
	}, nil
}

// CreateNote creates a new class note
func (s *classNoteServiceImpl) CreateNote(ctx context.Context, req *dto.CreateClassNoteRequest, file *multipart.FileHeader) (*dto.ClassNoteResponse, error) {
	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Create note model
	note := &models.ClassNote{
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Description:  req.Description,
		DepartmentID: req.DepartmentID,
		InstructorID: userID,
	}

	// Create note in database
	id, err := s.classNoteRepo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("error creating class note: %w", err)
	}

	// Convert to response DTO
	return &dto.ClassNoteResponse{
		ID:           id,
		CourseCode:   note.CourseCode,
		Title:        note.Title,
		Description:  note.Description,
		FileID:       note.FileID,
		DepartmentID: note.DepartmentID,
		InstructorID: note.InstructorID,
	}, nil
}

// UpdateNote updates an existing class note
func (s *classNoteServiceImpl) UpdateNote(ctx context.Context, id int64, req *dto.UpdateClassNoteRequest) (*dto.ClassNoteResponse, error) {
	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Validate ownership
	if err := s.authzService.ValidateClassNoteOwnership(ctx, id, userID); err != nil {
		return nil, err
	}

	// Get existing note
	note, err := s.classNoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting class note: %w", err)
	}
	if note == nil {
		return nil, apperrors.ErrClassNoteNotFound
	}

	// Update note fields
	note.CourseCode = req.CourseCode
	note.Title = req.Title
	note.Description = req.Description

	// Update note in database
	if err := s.classNoteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("error updating class note: %w", err)
	}

	// Convert to response DTO
	return &dto.ClassNoteResponse{
		ID:           note.ID,
		CourseCode:   note.CourseCode,
		Title:        note.Title,
		Description:  note.Description,
		FileID:       note.FileID,
		DepartmentID: note.DepartmentID,
		InstructorID: note.InstructorID,
	}, nil
}

// DeleteNote deletes a class note
func (s *classNoteServiceImpl) DeleteNote(ctx context.Context, id int64) error {
	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	// Validate ownership
	if err := s.authzService.ValidateClassNoteOwnership(ctx, id, userID); err != nil {
		return err
	}

	// Delete note from database
	if err := s.classNoteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting class note: %w", err)
	}

	return nil
}
