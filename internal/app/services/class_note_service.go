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
	"github.com/yigit/unisphere/internal/pkg/helpers"
)

// ClassNoteService defines the interface for class note operations
type ClassNoteService interface {
	GetAllNotes(ctx context.Context, filter *dto.ClassNoteFilterRequest) (*dto.ClassNoteListResponse, error)
	GetNoteByID(ctx context.Context, id int64) (*dto.ClassNoteResponse, error)
	CreateNote(ctx context.Context, req *dto.CreateClassNoteRequest, files []*multipart.FileHeader) (*dto.ClassNoteResponse, error)
	UpdateNote(ctx context.Context, id int64, req *dto.UpdateClassNoteRequest) (*dto.ClassNoteResponse, error)
	DeleteNote(ctx context.Context, id int64) error
	AddFileToNote(ctx context.Context, noteID int64, file *multipart.FileHeader) error
	AddFilesToNote(ctx context.Context, noteID int64, files []*multipart.FileHeader) (*dto.ClassNoteResponse, error)
	RemoveFileFromNote(ctx context.Context, noteID int64, fileID int64) error
	DeleteFileFromNote(ctx context.Context, noteID int64, fileID int64) error
	GetFileDetails(ctx context.Context, fileID int64) (*dto.ClassNoteFileResponse, error)
}

// classNoteServiceImpl implements ClassNoteService
type classNoteServiceImpl struct {
	classNoteRepo  *repositories.ClassNoteRepository
	departmentRepo *repositories.DepartmentRepository
	fileRepo       *repositories.FileRepository
	fileStorage    *filestorage.LocalStorage
	authzService   *auth.AuthorizationService
	logger         zerolog.Logger
}

// NewClassNoteService creates a new ClassNoteService
func NewClassNoteService(
	classNoteRepo *repositories.ClassNoteRepository,
	departmentRepo *repositories.DepartmentRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	authzService *auth.AuthorizationService,
	logger zerolog.Logger,
) ClassNoteService {
	return &classNoteServiceImpl{
		classNoteRepo:  classNoteRepo,
		departmentRepo: departmentRepo,
		fileRepo:       fileRepo,
		fileStorage:    fileStorage,
		authzService:   authzService,
		logger:         logger,
	}
}

// GetAllNotes retrieves all class notes with filtering, sorting and pagination
func (s *classNoteServiceImpl) GetAllNotes(ctx context.Context, filter *dto.ClassNoteFilterRequest) (*dto.ClassNoteListResponse, error) {
	s.logger.Debug().
		Interface("filter", filter).
		Msg("Getting all class notes")

	// Get notes from repository with sorting parameters
	notes, total, err := s.classNoteRepo.GetAll(ctx, filter.DepartmentID, filter.CourseCode, filter.InstructorID, 
		filter.Page, filter.PageSize, filter.SortBy, filter.SortOrder)
	if err != nil {
		s.logger.Error().Err(err).
			Interface("filter", filter).
			Msg("Failed to get class notes from repository")
		return nil, fmt.Errorf("error getting class notes: %w", err)
	}

	s.logger.Debug().
		Int("count", len(notes)).
		Int64("total", total).
		Str("sortBy", filter.SortBy).
		Str("sortOrder", filter.SortOrder).
		Msg("Retrieved class notes successfully")

	// Convert to response DTOs
	var noteResponses []dto.ClassNoteResponse
	for _, note := range notes {
		// Dosyaları getir
		files, err := s.classNoteRepo.GetClassNoteFiles(ctx, note.ID)
		if err != nil {
			s.logger.Error().Err(err).
				Int64("noteID", note.ID).
				Msg("Failed to get files for class note")
			// Hata durumunda boş dosya listesi ile devam edelim
			files = []*models.File{}
		}

		// Sadece dosya ID'lerini içeren yanıtlar oluştur
		var fileResponses []dto.SimpleClassNoteFileResponse
		for _, file := range files {
			fileResponses = append(fileResponses, dto.SimpleClassNoteFileResponse{
				ID: file.ID,
			})
		}

		noteResponses = append(noteResponses, dto.ClassNoteResponse{
			ID:           note.ID,
			CourseCode:   note.CourseCode,
			Title:        note.Title,
			Description:  note.Description,
			Content:      note.Content,
			DepartmentID: note.DepartmentID,
			UserID:       note.UserID,
			CreatedAt:    note.CreatedAt,
			UpdatedAt:    note.UpdatedAt,
			Files:        fileResponses,
		})
	}

	// Create response with pagination using the helper function
	paginationInfo := helpers.NewPaginationInfo(total, filter.Page, filter.PageSize)

	return &dto.ClassNoteListResponse{
		ClassNotes:     noteResponses,
		PaginationInfo: paginationInfo,
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

	// Sadece dosya ID'lerini içeren yanıtlar oluştur
	var fileResponses []dto.SimpleClassNoteFileResponse
	for _, file := range note.Files {
		fileResponses = append(fileResponses, dto.SimpleClassNoteFileResponse{
			ID: file.ID,
		})
	}

	// Convert to response DTO
	return &dto.ClassNoteResponse{
		ID:           note.ID,
		CourseCode:   note.CourseCode,
		Title:        note.Title,
		Description:  note.Description,
		Content:      note.Content,
		DepartmentID: note.DepartmentID,
		UserID:       note.UserID,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		Files:        fileResponses,
	}, nil
}

// GetFileDetails retrieves detailed information for a specific file
func (s *classNoteServiceImpl) GetFileDetails(ctx context.Context, fileID int64) (*dto.ClassNoteFileResponse, error) {
	// Get file from repository
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("fileID", fileID).
			Msg("Failed to get file details")
		return nil, fmt.Errorf("error getting file details: %w", err)
	}

	if file == nil {
		return nil, apperrors.ErrResourceNotFound
	}

	// Convert to response DTO
	return &dto.ClassNoteFileResponse{
		ID:        file.ID,
		FileName:  file.FileName,
		FileURL:   file.FileURL,
		FileSize:  file.FileSize,
		FileType:  file.FileType,
		CreatedAt: file.CreatedAt,
	}, nil
}

// CreateNote creates a new class note
func (s *classNoteServiceImpl) CreateNote(ctx context.Context, req *dto.CreateClassNoteRequest, files []*multipart.FileHeader) (*dto.ClassNoteResponse, error) {
	s.logger.Debug().
		Interface("request", req).
		Int("fileCount", len(files)).
		Msg("Creating new class note")

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Create note model
	note := &models.ClassNote{
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Description:  req.Description,
		Content:      req.Content,
		DepartmentID: req.DepartmentID,
		UserID:       userID,
	}

	// Save note to database
	noteID, err := s.classNoteRepo.Create(ctx, note)
	if err != nil {
		s.logger.Error().Err(err).
			Interface("note", note).
			Msg("Failed to create class note")
		return nil, fmt.Errorf("failed to create class note: %w", err)
	}

	note.ID = noteID

	s.logger.Debug().
		Int64("noteID", noteID).
		Msg("Class note created successfully")

	// Sadece dosya ID'lerini içeren yanıtlar oluştur
	var fileResponses []dto.SimpleClassNoteFileResponse

	// Process files if any
	if len(files) > 0 {
		for _, file := range files {
			// Dosyayı kaydet
			fileURL, err := s.fileStorage.SaveFileWithPath(file, "class_notes")
			if err != nil {
				s.logger.Error().Err(err).
					Str("fileName", file.Filename).
					Msg("Failed to save file")
				continue
			}

			s.logger.Debug().
				Str("fileURL", fileURL).
				Msg("File saved successfully")

			// Extract relative path from URL
			relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
			relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")

			// Create file record
			fileRecord := &models.File{
				FileName:     file.Filename,
				FilePath:     relativeFilePath,
				FileURL:      fileURL,
				FileSize:     file.Size,
				FileType:     file.Header.Get("Content-Type"),
				ResourceType: "CLASS_NOTE",
				ResourceID:   noteID,
				UploadedBy:   userID,
			}

			// Save file record to database
			fileID, err := s.fileRepo.Create(ctx, fileRecord)
			if err != nil {
				s.logger.Error().Err(err).
					Interface("fileRecord", fileRecord).
					Msg("Failed to save file record")
				// Fiziksel dosyayı sil
				_ = s.fileStorage.DeleteFile(fileRecord.FilePath)
				continue
			}

			fileRecord.ID = fileID

			// Dosya ve not arasında ilişki kur
			err = s.classNoteRepo.AddFileToClassNote(ctx, noteID, fileID)
			if err != nil {
				s.logger.Error().Err(err).
					Int64("noteID", noteID).
					Int64("fileID", fileID).
					Msg("Failed to add file to class note")
				continue
			}

			// Sadece dosya ID'sini ekle
			fileResponses = append(fileResponses, dto.SimpleClassNoteFileResponse{
				ID: fileID,
			})
		}
	}

	// Return response
	return &dto.ClassNoteResponse{
		ID:           noteID,
		CourseCode:   note.CourseCode,
		Title:        note.Title,
		Description:  note.Description,
		Content:      note.Content,
		DepartmentID: note.DepartmentID,
		UserID:       note.UserID,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		Files:        fileResponses,
	}, nil
}

// UpdateNote updates an existing class note
func (s *classNoteServiceImpl) UpdateNote(ctx context.Context, id int64, req *dto.UpdateClassNoteRequest) (*dto.ClassNoteResponse, error) {
	s.logger.Debug().
		Int64("id", id).
		Interface("request", req).
		Msg("Updating class note")

	// Get note from repository
	existingNote, err := s.classNoteRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("id", id).
			Msg("Error getting class note for update")
		return nil, fmt.Errorf("error getting class note: %w", err)
	}
	if existingNote == nil {
		s.logger.Error().
			Int64("id", id).
			Msg("Class note not found for update")
		return nil, apperrors.ErrClassNoteNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is authorized to update note
	if err := s.authzService.ValidateClassNoteOwnership(ctx, existingNote.ID, userID); err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Int64("noteID", existingNote.ID).
			Msg("User not authorized to update note")
		return nil, err
	}

	// Update note model
	note := &models.ClassNote{
		ID:           id,
		CourseCode:   req.CourseCode,
		Title:        req.Title,
		Description:  req.Description,
		Content:      req.Content,
		DepartmentID: existingNote.DepartmentID, // Keep the same department
		UserID:       existingNote.UserID,       // Keep the same user
	}

	s.logger.Debug().
		Interface("noteToUpdate", note).
		Msg("Calling repository Update method")

	// Update note in database
	if err := s.classNoteRepo.Update(ctx, note); err != nil {
		s.logger.Error().Err(err).
			Int64("id", id).
			Msg("Error updating class note in database")
		return nil, fmt.Errorf("error updating class note: %w", err)
	}

	s.logger.Debug().
		Int64("id", id).
		Msg("Class note updated successfully, fetching updated note")

	// Get updated note
	updatedNote, err := s.classNoteRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("id", id).
			Msg("Error getting updated class note")
		return nil, fmt.Errorf("error getting updated class note: %w", err)
	}

	// Sadece dosya ID'lerini içeren yanıtlar oluştur
	var fileResponses []dto.SimpleClassNoteFileResponse
	for _, file := range updatedNote.Files {
		fileResponses = append(fileResponses, dto.SimpleClassNoteFileResponse{
			ID: file.ID,
		})
	}

	s.logger.Debug().
		Int64("id", updatedNote.ID).
		Str("courseCode", updatedNote.CourseCode).
		Str("title", updatedNote.Title).
		Int("fileCount", len(fileResponses)).
		Msg("Returning updated class note")

	// Convert to response DTO
	return &dto.ClassNoteResponse{
		ID:           updatedNote.ID,
		CourseCode:   updatedNote.CourseCode,
		Title:        updatedNote.Title,
		Description:  updatedNote.Description,
		Content:      updatedNote.Content,
		DepartmentID: updatedNote.DepartmentID,
		UserID:       updatedNote.UserID,
		CreatedAt:    updatedNote.CreatedAt,
		UpdatedAt:    updatedNote.UpdatedAt,
		Files:        fileResponses,
	}, nil
}

// DeleteNote deletes a class note
func (s *classNoteServiceImpl) DeleteNote(ctx context.Context, id int64) error {
	// Get existing note
	existingNote, err := s.classNoteRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting class note: %w", err)
	}
	if existingNote == nil {
		return apperrors.ErrClassNoteNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user is authorized to delete the note
	if err := s.authzService.ValidateClassNoteOwnership(ctx, id, userID); err != nil {
		return err
	}

	// Delete all associated files
	for _, file := range existingNote.Files {
		// Delete physical file
		err := s.fileStorage.DeleteFile(file.FilePath)
		if err != nil {
			s.logger.Warn().Err(err).
				Int64("fileID", file.ID).
				Str("filePath", file.FilePath).
				Msg("Failed to delete physical file")
		}

		// Removing from class_note_files handled by foreign key cascade

		// Delete file record
		err = s.fileRepo.Delete(ctx, file.ID)
		if err != nil {
			s.logger.Warn().Err(err).
				Int64("fileID", file.ID).
				Msg("Failed to delete file record")
		}
	}

	// Delete note
	err = s.classNoteRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting class note: %w", err)
	}

	return nil
}

// AddFileToNote adds a file to an existing class note
func (s *classNoteServiceImpl) AddFileToNote(ctx context.Context, noteID int64, file *multipart.FileHeader) error {
	// Get existing note
	existingNote, err := s.classNoteRepo.GetByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("error getting class note: %w", err)
	}
	if existingNote == nil {
		return apperrors.ErrClassNoteNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to update
	// Only the creator can update the note
	if existingNote.UserID != userID {
		return fmt.Errorf("unauthorized: only the creator can update this note")
	}

	// Save file
	fileURL, err := s.fileStorage.SaveFileWithPath(file, "class_notes")
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Extract relative path from URL
	relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
	relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")

	// Create file record
	fileRecord := &models.File{
		FileName:     file.Filename,
		FilePath:     relativeFilePath,
		FileURL:      fileURL,
		FileSize:     file.Size,
		FileType:     file.Header.Get("Content-Type"),
		ResourceType: "CLASS_NOTE",
		ResourceID:   noteID,
		UploadedBy:   userID,
	}

	// Save file record to database
	fileID, err := s.fileRepo.Create(ctx, fileRecord)
	if err != nil {
		// If DB save fails, delete the physical file
		_ = s.fileStorage.DeleteFile(fileRecord.FilePath)
		return fmt.Errorf("failed to save file record: %w", err)
	}

	// Add file to class note
	err = s.classNoteRepo.AddFileToClassNote(ctx, noteID, fileID)
	if err != nil {
		// If relation fails, delete file and file record
		_ = s.fileStorage.DeleteFile(fileRecord.FilePath)
		_ = s.fileRepo.Delete(ctx, fileID)
		return fmt.Errorf("failed to add file to class note: %w", err)
	}

	return nil
}

// AddFilesToNote adds multiple files to an existing class note
func (s *classNoteServiceImpl) AddFilesToNote(ctx context.Context, noteID int64, files []*multipart.FileHeader) (*dto.ClassNoteResponse, error) {
	s.logger.Debug().
		Int64("noteID", noteID).
		Int("fileCount", len(files)).
		Msg("Adding files to class note")

	// Get note from repository to check if it exists
	note, err := s.classNoteRepo.GetByID(ctx, noteID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("noteID", noteID).
			Msg("Failed to get class note")
		return nil, fmt.Errorf("failed to get class note: %w", err)
	}
	if note == nil {
		s.logger.Error().
			Int64("noteID", noteID).
			Msg("Class note not found")
		return nil, apperrors.ErrClassNoteNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().
			Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is authorized to manage this note
	if err := s.authzService.ValidateClassNoteOwnership(ctx, noteID, userID); err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Int64("noteID", noteID).
			Msg("User not authorized to manage class note")
		return nil, err
	}

	// Upload files one by one
	fileUploadErrors := []error{}
	for _, file := range files {
		err := s.AddFileToNote(ctx, noteID, file)
		if err != nil {
			s.logger.Error().Err(err).
				Str("fileName", file.Filename).
				Int64("noteID", noteID).
				Msg("Failed to add file to class note")
			fileUploadErrors = append(fileUploadErrors, err)
		}
	}

	// Check if any files were uploaded
	if len(fileUploadErrors) == len(files) && len(files) > 0 {
		// All file uploads failed
		s.logger.Error().
			Int("totalFiles", len(files)).
			Int("errorCount", len(fileUploadErrors)).
			Msg("All file uploads failed")
		return nil, fmt.Errorf("all file uploads failed")
	}

	// Get the updated note
	updatedNote, err := s.classNoteRepo.GetByID(ctx, noteID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("noteID", noteID).
			Msg("Failed to get updated class note")
		return nil, fmt.Errorf("failed to get updated class note: %w", err)
	}

	// Sadece dosya ID'lerini içeren yanıtlar oluştur
	var fileResponses []dto.SimpleClassNoteFileResponse
	for _, file := range updatedNote.Files {
		fileResponses = append(fileResponses, dto.SimpleClassNoteFileResponse{
			ID: file.ID,
		})
	}

	// Return the updated note
	return &dto.ClassNoteResponse{
		ID:           updatedNote.ID,
		CourseCode:   updatedNote.CourseCode,
		Title:        updatedNote.Title,
		Description:  updatedNote.Description,
		Content:      updatedNote.Content,
		DepartmentID: updatedNote.DepartmentID,
		UserID:       updatedNote.UserID,
		CreatedAt:    updatedNote.CreatedAt,
		UpdatedAt:    updatedNote.UpdatedAt,
		Files:        fileResponses,
	}, nil
}

// RemoveFileFromNote removes a file from a class note
func (s *classNoteServiceImpl) RemoveFileFromNote(ctx context.Context, noteID int64, fileID int64) error {
	// Get existing note
	existingNote, err := s.classNoteRepo.GetByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("error getting class note: %w", err)
	}
	if existingNote == nil {
		return apperrors.ErrClassNoteNotFound
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user has permission to update
	// Only the creator can update the note
	if existingNote.UserID != userID {
		return fmt.Errorf("unauthorized: only the creator can update this note")
	}

	// Get file
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("error getting file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// Ensure file belongs to this note
	var fileFound bool
	for _, f := range existingNote.Files {
		if f.ID == fileID {
			fileFound = true
			break
		}
	}
	if !fileFound {
		return fmt.Errorf("file does not belong to this note")
	}

	// Remove file from class note
	err = s.classNoteRepo.RemoveFileFromClassNote(ctx, noteID, fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file from class note: %w", err)
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

// DeleteFileFromNote deletes a file from a class note (alias for RemoveFileFromNote)
func (s *classNoteServiceImpl) DeleteFileFromNote(ctx context.Context, noteID int64, fileID int64) error {
	// This method is an alias for RemoveFileFromNote to maintain backward compatibility
	return s.RemoveFileFromNote(ctx, noteID, fileID)
}
