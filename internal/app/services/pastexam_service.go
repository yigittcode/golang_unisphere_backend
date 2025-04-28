package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// PastExamService defines the interface for past exam operations
type PastExamService interface {
	GetAllExams(ctx context.Context, filter *dto.PastExamFilterRequest) (*dto.PastExamListResponse, error)
	GetExamByID(ctx context.Context, id int64) (*dto.PastExamResponse, error)
	CreateExam(ctx context.Context, req *dto.CreatePastExamRequest, file *multipart.FileHeader) (*dto.PastExamResponse, error)
	UpdateExam(ctx context.Context, id int64, req *dto.UpdatePastExamRequest) (*dto.PastExamResponse, error)
	DeleteExam(ctx context.Context, id int64) error
}

// pastExamServiceImpl implements PastExamService
type pastExamServiceImpl struct {
	pastExamRepo   *repositories.PastExamRepository
	departmentRepo *repositories.DepartmentRepository
	authzService   *auth.AuthorizationService
}

// NewPastExamService creates a new PastExamService
func NewPastExamService(
	pastExamRepo *repositories.PastExamRepository,
	departmentRepo *repositories.DepartmentRepository,
	authzService *auth.AuthorizationService,
) PastExamService {
	return &pastExamServiceImpl{
		pastExamRepo:   pastExamRepo,
		departmentRepo: departmentRepo,
		authzService:   authzService,
	}
}

// GetAllExams retrieves all past exams with filtering and pagination
func (s *pastExamServiceImpl) GetAllExams(ctx context.Context, filter *dto.PastExamFilterRequest) (*dto.PastExamListResponse, error) {
	// Get exams from repository
	exams, total, err := s.pastExamRepo.GetAll(ctx, filter.DepartmentID, filter.CourseCode, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("error getting past exams: %w", err)
	}

	// Convert to response DTOs
	var examResponses []dto.PastExamResponse
	for _, exam := range exams {
		examResponses = append(examResponses, dto.PastExamResponse{
			ID:           exam.ID,
			CourseCode:   exam.CourseCode,
			Year:         exam.Year,
			Term:         exam.Term,
			FileID:       exam.FileID,
			DepartmentID: exam.DepartmentID,
			InstructorID: exam.InstructorID,
		})
	}

	// Create response with pagination
	totalPages := (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)
	return &dto.PastExamListResponse{
		PastExams: examResponses,
		PaginationInfo: dto.PaginationInfo{
			CurrentPage: filter.Page,
			PageSize:    filter.PageSize,
			TotalItems:  total,
			TotalPages:  int(totalPages),
		},
	}, nil
}

// GetExamByID retrieves a past exam by ID
func (s *pastExamServiceImpl) GetExamByID(ctx context.Context, id int64) (*dto.PastExamResponse, error) {
	// Get exam from repository
	exam, err := s.pastExamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting past exam: %w", err)
	}
	if exam == nil {
		return nil, apperrors.ErrPastExamNotFound
	}

	// Convert to response DTO
	return &dto.PastExamResponse{
		ID:           exam.ID,
		CourseCode:   exam.CourseCode,
		Year:         exam.Year,
		Term:         exam.Term,
		FileID:       exam.FileID,
		DepartmentID: exam.DepartmentID,
		InstructorID: exam.InstructorID,
	}, nil
}

// CreateExam creates a new past exam
func (s *pastExamServiceImpl) CreateExam(ctx context.Context, req *dto.CreatePastExamRequest, file *multipart.FileHeader) (*dto.PastExamResponse, error) {
	// TODO: Implement file upload and exam creation
	return nil, nil
}

// UpdateExam updates an existing past exam
func (s *pastExamServiceImpl) UpdateExam(ctx context.Context, id int64, req *dto.UpdatePastExamRequest) (*dto.PastExamResponse, error) {
	// TODO: Implement exam update
	return nil, nil
}

// DeleteExam deletes a past exam
func (s *pastExamServiceImpl) DeleteExam(ctx context.Context, id int64) error {
	// TODO: Implement exam deletion
	return nil
}
