package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
)

// PastExamService defines the interface for past exam operations
type PastExamService interface {
	GetAllExams(ctx context.Context, filter *dto.PastExamFilterRequest) (*dto.PastExamListResponse, error)
	GetExamByID(ctx context.Context, id int64) (*dto.PastExamResponse, error)
	CreateExam(ctx context.Context, req *dto.CreatePastExamRequest, files []*multipart.FileHeader) (*dto.PastExamResponse, error)
	UpdateExam(ctx context.Context, id int64, req *dto.UpdatePastExamRequest) (*dto.PastExamResponse, error)
	DeleteExam(ctx context.Context, id int64) error
	AddFileToPastExam(ctx context.Context, examID int64, file *multipart.FileHeader) error
	RemoveFileFromPastExam(ctx context.Context, examID int64, fileID int64) error
}

// pastExamServiceImpl implements PastExamService
type pastExamServiceImpl struct {
	pastExamRepo   *repositories.PastExamRepository
	departmentRepo *repositories.DepartmentRepository
	fileRepo       *repositories.FileRepository
	fileStorage    *filestorage.LocalStorage
	authzService   *auth.AuthorizationService
	logger         zerolog.Logger
}

// NewPastExamService creates a new PastExamService
func NewPastExamService(
	pastExamRepo *repositories.PastExamRepository,
	departmentRepo *repositories.DepartmentRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	authzService *auth.AuthorizationService,
	logger zerolog.Logger,
) PastExamService {
	return &pastExamServiceImpl{
		pastExamRepo:   pastExamRepo,
		departmentRepo: departmentRepo,
		fileRepo:       fileRepo,
		fileStorage:    fileStorage,
		authzService:   authzService,
		logger:         logger,
	}
}

// GetAllExams retrieves all past exams with filtering and pagination
func (s *pastExamServiceImpl) GetAllExams(ctx context.Context, filter *dto.PastExamFilterRequest) (*dto.PastExamListResponse, error) {
	// Get exams from repository
	exams, total, err := s.pastExamRepo.GetAll(ctx, filter.DepartmentID, filter.CourseCode, filter.Year, filter.Term, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("error getting past exams: %w", err)
	}

	// Convert to response DTOs
	var examResponses []dto.PastExamResponse
	for _, exam := range exams {
		// Dosya yanıtlarını oluştur
		var fileResponses []dto.PastExamFileResponse
		for _, file := range exam.Files {
			fileResponses = append(fileResponses, dto.PastExamFileResponse{
				ID:        file.ID,
				FileName:  file.FileName,
				FileURL:   file.FileURL,
				FileSize:  file.FileSize,
				FileType:  file.FileType,
				CreatedAt: file.CreatedAt,
			})
		}

		examResponses = append(examResponses, dto.PastExamResponse{
			ID:           exam.ID,
			CourseCode:   exam.CourseCode,
			Year:         exam.Year,
			Term:         string(exam.Term),
			Title:        exam.Title,
			Content:      exam.Content,
			DepartmentID: exam.DepartmentID,
			InstructorID: exam.InstructorID,
			Files:        fileResponses,
			CreatedAt:    exam.CreatedAt,
			UpdatedAt:    exam.UpdatedAt,
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

	// Dosya yanıtlarını oluştur
	var fileResponses []dto.PastExamFileResponse
	for _, file := range exam.Files {
		fileResponses = append(fileResponses, dto.PastExamFileResponse{
			ID:        file.ID,
			FileName:  file.FileName,
			FileURL:   file.FileURL,
			FileSize:  file.FileSize,
			FileType:  file.FileType,
			CreatedAt: file.CreatedAt,
		})
	}

	// Convert to response DTO
	return &dto.PastExamResponse{
		ID:           exam.ID,
		CourseCode:   exam.CourseCode,
		Year:         exam.Year,
		Term:         string(exam.Term),
		Title:        exam.Title,
		Content:      exam.Content,
		DepartmentID: exam.DepartmentID,
		InstructorID: exam.InstructorID,
		Files:        fileResponses,
		CreatedAt:    exam.CreatedAt,
		UpdatedAt:    exam.UpdatedAt,
	}, nil
}

// CreateExam creates a new past exam
func (s *pastExamServiceImpl) CreateExam(ctx context.Context, req *dto.CreatePastExamRequest, files []*multipart.FileHeader) (*dto.PastExamResponse, error) {
	s.logger.Debug().
		Interface("request", req).
		Int("fileCount", len(files)).
		Msg("Creating new past exam")

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Create exam model
	exam := &models.PastExam{
		CourseCode:   req.CourseCode,
		Year:         req.Year,
		Term:         models.Term(req.Term),
		Title:        req.Title,
		Content:      req.Content,
		DepartmentID: req.DepartmentID,
		InstructorID: userID, // Use current user as instructor
	}

	// Save exam to database
	examID, err := s.pastExamRepo.Create(ctx, exam)
	if err != nil {
		s.logger.Error().Err(err).
			Interface("exam", exam).
			Msg("Failed to create past exam")
		return nil, fmt.Errorf("failed to create past exam: %w", err)
	}

	exam.ID = examID

	// Process uploaded files
	var savedFiles []*models.File
	for _, fileHeader := range files {
		// Upload file to storage
		file, err := s.uploadFile(ctx, fileHeader, models.FileTypePastExam, examID, userID)
		if err != nil {
			s.logger.Error().Err(err).
				Str("filename", fileHeader.Filename).
				Int64("examID", examID).
				Msg("Failed to upload file for past exam")
			continue
		}

		// Link file to past exam
		err = s.pastExamRepo.AddFileToPastExam(ctx, examID, file.ID)
		if err != nil {
			s.logger.Error().Err(err).
				Int64("fileID", file.ID).
				Int64("examID", examID).
				Msg("Failed to link file to past exam")
			continue
		}

		savedFiles = append(savedFiles, file)
	}

	// Add files to exam model
	exam.Files = savedFiles

	// Convert to response DTO
	var fileResponses []dto.PastExamFileResponse
	for _, file := range savedFiles {
		fileResponses = append(fileResponses, dto.PastExamFileResponse{
			ID:        file.ID,
			FileName:  file.FileName,
			FileURL:   file.FileURL,
			FileSize:  file.FileSize,
			FileType:  file.FileType,
			CreatedAt: file.CreatedAt,
		})
	}

	return &dto.PastExamResponse{
		ID:           exam.ID,
		CourseCode:   exam.CourseCode,
		Year:         exam.Year,
		Term:         string(exam.Term),
		Title:        exam.Title,
		Content:      exam.Content,
		DepartmentID: exam.DepartmentID,
		InstructorID: exam.InstructorID,
		Files:        fileResponses,
		CreatedAt:    exam.CreatedAt,
		UpdatedAt:    exam.UpdatedAt,
	}, nil
}

// uploadFile uploads a file to storage and saves its metadata to the database
func (s *pastExamServiceImpl) uploadFile(ctx context.Context, fileHeader *multipart.FileHeader, resourceType models.FileType, resourceID int64, userID int64) (*models.File, error) {
	// Open the file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer src.Close()

	// Generate a storage path based on resource type and ID
	subPath := fmt.Sprintf("%s_%d", resourceType, resourceID)

	// Upload to storage
	fileURL, err := s.fileStorage.SaveFileWithPath(fileHeader, subPath)
	if err != nil {
		return nil, fmt.Errorf("error uploading file: %w", err)
	}

	// Extract relative path from URL
	relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
	relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")

	// Create file model
	file := &models.File{
		FileName:     fileHeader.Filename,
		FilePath:     relativeFilePath,
		FileURL:      fileURL,
		FileSize:     fileHeader.Size,
		FileType:     fileHeader.Header.Get("Content-Type"),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UploadedBy:   userID,
	}

	// Save file metadata to database
	fileID, err := s.fileRepo.Create(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("error saving file metadata: %w", err)
	}
	file.ID = fileID

	return file, nil
}

// UpdateExam updates an existing past exam
func (s *pastExamServiceImpl) UpdateExam(ctx context.Context, id int64, req *dto.UpdatePastExamRequest) (*dto.PastExamResponse, error) {
	// Get existing exam
	existingExam, err := s.pastExamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting past exam: %w", err)
	}
	if existingExam == nil {
		return nil, apperrors.ErrPastExamNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to update
	// Only the instructor who created the exam can update it
	if existingExam.InstructorID != userID {
		return nil, fmt.Errorf("unauthorized: only the creator can update this exam")
	}

	// Update exam model with new values
	updatedExam := &models.PastExam{
		ID:           id,
		CourseCode:   req.CourseCode,
		Year:         req.Year,
		Term:         models.Term(req.Term),
		Title:        req.Title,
		Content:      req.Content,
		DepartmentID: existingExam.DepartmentID, // Keep original department
		InstructorID: existingExam.InstructorID, // Keep original instructor
	}

	// Update exam in database
	err = s.pastExamRepo.Update(ctx, updatedExam)
	if err != nil {
		s.logger.Error().Err(err).
			Interface("exam", updatedExam).
			Msg("Failed to update past exam")
		return nil, fmt.Errorf("failed to update past exam: %w", err)
	}

	// Get updated exam with files
	updatedExamFull, err := s.pastExamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting updated past exam: %w", err)
	}

	// Convert to response DTO
	var fileResponses []dto.PastExamFileResponse
	for _, file := range updatedExamFull.Files {
		fileResponses = append(fileResponses, dto.PastExamFileResponse{
			ID:        file.ID,
			FileName:  file.FileName,
			FileURL:   file.FileURL,
			FileSize:  file.FileSize,
			FileType:  file.FileType,
			CreatedAt: file.CreatedAt,
		})
	}

	return &dto.PastExamResponse{
		ID:           updatedExamFull.ID,
		CourseCode:   updatedExamFull.CourseCode,
		Year:         updatedExamFull.Year,
		Term:         string(updatedExamFull.Term),
		Title:        updatedExamFull.Title,
		Content:      updatedExamFull.Content,
		DepartmentID: updatedExamFull.DepartmentID,
		InstructorID: updatedExamFull.InstructorID,
		Files:        fileResponses,
		CreatedAt:    updatedExamFull.CreatedAt,
		UpdatedAt:    updatedExamFull.UpdatedAt,
	}, nil
}

// DeleteExam deletes a past exam
func (s *pastExamServiceImpl) DeleteExam(ctx context.Context, id int64) error {
	// Get existing exam
	existingExam, err := s.pastExamRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting past exam: %w", err)
	}
	if existingExam == nil {
		return apperrors.ErrPastExamNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to delete
	// Only the instructor who created the exam can delete it
	if existingExam.InstructorID != userID {
		return fmt.Errorf("unauthorized: only the creator can delete this exam")
	}

	// Delete all associated files
	for _, file := range existingExam.Files {
		// Delete physical file
		err := s.fileStorage.DeleteFile(file.FilePath)
		if err != nil {
			s.logger.Warn().Err(err).
				Int64("fileID", file.ID).
				Str("filePath", file.FilePath).
				Msg("Failed to delete physical file")
		}

		// Removing from past_exam_files handled by foreign key cascade

		// Delete file record
		err = s.fileRepo.Delete(ctx, file.ID)
		if err != nil {
			s.logger.Warn().Err(err).
				Int64("fileID", file.ID).
				Msg("Failed to delete file record")
		}
	}

	// Delete exam
	err = s.pastExamRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting past exam: %w", err)
	}

	return nil
}

// AddFileToPastExam adds a file to an existing past exam
func (s *pastExamServiceImpl) AddFileToPastExam(ctx context.Context, examID int64, file *multipart.FileHeader) error {
	// Get existing exam
	existingExam, err := s.pastExamRepo.GetByID(ctx, examID)
	if err != nil {
		return fmt.Errorf("error getting past exam: %w", err)
	}
	if existingExam == nil {
		return apperrors.ErrPastExamNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to update
	// Only the instructor who created the exam can update it
	if existingExam.InstructorID != userID {
		return fmt.Errorf("unauthorized: only the creator can update this exam")
	}

	// Upload file
	uploadedFile, err := s.uploadFile(ctx, file, models.FileTypePastExam, examID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("filename", file.Filename).
			Int64("examID", examID).
			Msg("Failed to upload file for past exam")
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Link file to past exam
	err = s.pastExamRepo.AddFileToPastExam(ctx, examID, uploadedFile.ID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("fileID", uploadedFile.ID).
			Int64("examID", examID).
			Msg("Failed to link file to past exam")

		// Clean up - delete the file if we couldn't link it
		_ = s.fileStorage.DeleteFile(uploadedFile.FilePath)
		_ = s.fileRepo.Delete(ctx, uploadedFile.ID)

		return fmt.Errorf("failed to link file to past exam: %w", err)
	}

	return nil
}

// RemoveFileFromPastExam removes a file from a past exam
func (s *pastExamServiceImpl) RemoveFileFromPastExam(ctx context.Context, examID int64, fileID int64) error {
	// Get existing exam
	existingExam, err := s.pastExamRepo.GetByID(ctx, examID)
	if err != nil {
		return fmt.Errorf("error getting past exam: %w", err)
	}
	if existingExam == nil {
		return apperrors.ErrPastExamNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to update
	// Only the instructor who created the exam can update it
	if existingExam.InstructorID != userID {
		return fmt.Errorf("unauthorized: only the creator can update this exam")
	}

	// Get file
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// Ensure file belongs to this exam
	var fileFound bool
	for _, f := range existingExam.Files {
		if f.ID == fileID {
			fileFound = true
			break
		}
	}
	if !fileFound {
		return fmt.Errorf("file does not belong to this exam")
	}

	// Remove file from past exam
	err = s.pastExamRepo.RemoveFileFromPastExam(ctx, examID, fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file from past exam: %w", err)
	}

	// Delete file record
	err = s.fileRepo.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// Delete physical file
	err = s.fileStorage.DeleteFile(file.FilePath)
	if err != nil {
		s.logger.Warn().Err(err).
			Int64("fileID", fileID).
			Str("filePath", file.FilePath).
			Msg("Failed to delete physical file")
	}

	return nil
}
