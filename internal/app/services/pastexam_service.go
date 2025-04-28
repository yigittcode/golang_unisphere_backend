package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// PastExamService defines the interface for past exam related operations
type PastExamService interface {
	GetAllPastExams(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]models.PastExam, dto.PaginationInfo, error)
	GetPastExamByID(ctx context.Context, id int64) (*models.PastExam, error)
	CreatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64) (int64, error)
	UpdatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64, newFilePath *string) error
	DeletePastExam(ctx context.Context, id int64, userID int64) error
	AddFileToPastExam(ctx context.Context, examID int64, file *models.File) (int64, error)
	RemoveFileFromPastExam(ctx context.Context, examID, fileID, userID int64) error
	GetPastExamFiles(ctx context.Context, examID int64) ([]*models.File, error)
}

// pastExamServiceImpl implements the PastExamService interface
type pastExamServiceImpl struct {
	pastExamRepo *repositories.PastExamRepository
	authService  *auth.AuthorizationService
}

// NewPastExamService creates a new past exam service instance
func NewPastExamService(pastExamRepo *repositories.PastExamRepository, authService *auth.AuthorizationService) PastExamService {
	return &pastExamServiceImpl{
		pastExamRepo: pastExamRepo,
		authService:  authService,
	}
}

// validatePastExam performs basic validation on past exam fields
func (s *pastExamServiceImpl) validatePastExam(pastExam *models.PastExam) error {
	if pastExam == nil {
		return fmt.Errorf("%w: past exam data cannot be nil", apperrors.ErrValidationFailed)
	}

	// Validate year
	currentYear := time.Now().Year()
	if pastExam.Year < 1900 || pastExam.Year > currentYear+1 { // Allow one year in the future
		return fmt.Errorf("%w: year must be between 1900 and %d", apperrors.ErrValidationFailed, currentYear+1)
	}

	// Validate term
	term := strings.ToUpper(string(pastExam.Term))
	if term != "FALL" && term != "SPRING" {
		return fmt.Errorf("%w: term must be FALL or SPRING", apperrors.ErrValidationFailed)
	}

	// Validate department ID
	if pastExam.DepartmentID <= 0 {
		return fmt.Errorf("%w: department ID must be positive", apperrors.ErrValidationFailed)
	}

	// Validate course code
	if strings.TrimSpace(pastExam.CourseCode) == "" {
		return fmt.Errorf("%w: course code cannot be empty", apperrors.ErrValidationFailed)
	}

	// Validate title
	if strings.TrimSpace(pastExam.Title) == "" {
		return fmt.Errorf("%w: title cannot be empty", apperrors.ErrValidationFailed)
	}

	// Validate content
	if strings.TrimSpace(pastExam.Content) == "" {
		return fmt.Errorf("%w: content cannot be empty", apperrors.ErrValidationFailed)
	}

	return nil
}

// CreatePastExam creates a new past exam, only accessible to instructors
func (s *pastExamServiceImpl) CreatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64) (int64, error) {
	// Validate past exam data
	if err := s.validatePastExam(pastExam); err != nil {
		return 0, err
	}

	// Validate that the user is an instructor and get instructor ID
	instructor, err := s.authService.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrNotInstructor) {
			return 0, apperrors.ErrInstructorOnly
		}
		return 0, fmt.Errorf("user validation error: %w", err)
	}

	// Set instructor information
	pastExam.InstructorID = instructor.ID

	// Get user information and set in the exam for display purposes
	user, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Set uploader information for display purposes
	pastExam.UploadedByName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	pastExam.UploadedByEmail = user.Email

	// Set timestamps
	now := time.Now()
	pastExam.CreatedAt = now
	pastExam.UpdatedAt = now

	id, err := s.pastExamRepo.CreatePastExam(ctx, pastExam)
	if err != nil {
		return 0, fmt.Errorf("error creating past exam: %w", err)
	}

	return id, nil
}

// GetPastExamByID retrieves a past exam by ID
func (s *pastExamServiceImpl) GetPastExamByID(ctx context.Context, id int64) (*models.PastExam, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid ID", apperrors.ErrValidationFailed)
	}

	pastExam, err := s.pastExamRepo.GetPastExamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return nil, apperrors.ErrPastExamNotFound
		}
		return nil, fmt.Errorf("error retrieving past exam: %w", err)
	}

	return pastExam, nil
}

// GetAllPastExams retrieves all past exams with pagination and filtering
func (s *pastExamServiceImpl) GetAllPastExams(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]models.PastExam, dto.PaginationInfo, error) {
	// Validate pagination parameters using helpers
	if page < helpers.DefaultPage {
		page = helpers.DefaultPage
	}
	if pageSize <= 0 || pageSize > helpers.MaxPageSize {
		pageSize = helpers.DefaultPageSize
	}

	// Validate filter values if present
	if filters != nil {
		// Year validation
		if yearVal, ok := filters["year"]; ok && yearVal != nil {
			if year, ok := yearVal.(int); ok {
				currentYear := time.Now().Year()
				if year < 1900 || year > currentYear+1 {
					return nil, dto.PaginationInfo{}, fmt.Errorf("%w: year must be between 1900 and %d", apperrors.ErrValidationFailed, currentYear+1)
				}
			}
		}

		// Term validation
		if termVal, ok := filters["term"]; ok && termVal != nil {
			if term, ok := termVal.(string); ok {
				termUpper := strings.ToUpper(term)
				if termUpper != "FALL" && termUpper != "SPRING" {
					return nil, dto.PaginationInfo{}, fmt.Errorf("%w: term must be FALL or SPRING", apperrors.ErrValidationFailed)
				}
				filters["term"] = termUpper // Normalize the term to uppercase
			}
		}
	}

	// Call repository (page is 0-based)
	pastExams, paginationInfo, err := s.pastExamRepo.GetAllPastExams(ctx, page, pageSize, filters)
	if err != nil {
		return nil, dto.PaginationInfo{}, fmt.Errorf("error retrieving past exams: %w", err) // Return empty pagination on error
	}

	return pastExams, paginationInfo, nil // Return pagination info directly from repo
}

// UpdatePastExam updates an existing past exam.
// If a new file path is provided, it updates the FileURL.
func (s *pastExamServiceImpl) UpdatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64, newFilePath *string) error {
	// Validate past exam data (excluding FileURL from the model, as it might be replaced)
	if err := s.validatePastExam(pastExam); err != nil {
		return err
	}

	// Validate exam ID
	if pastExam.ID <= 0 {
		return fmt.Errorf("%w: invalid exam ID", apperrors.ErrValidationFailed)
	}

	// Check if the exam exists and get current data
	existingExam, err := s.pastExamRepo.GetPastExamByID(ctx, pastExam.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return apperrors.ErrPastExamNotFound
		}
		return fmt.Errorf("error checking exam existence: %w", err)
	}

	// Check if the user can modify this exam
	err = s.authService.ValidatePastExamOwnership(ctx, pastExam.ID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return apperrors.ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	// Preserve original instructor information
	pastExam.InstructorID = existingExam.InstructorID

	// Preserve display information (although these are not persisted)
	pastExam.UploadedByName = existingExam.UploadedByName
	pastExam.UploadedByEmail = existingExam.UploadedByEmail
	// Preserve created_at
	pastExam.CreatedAt = existingExam.CreatedAt

	// Handle FileURL update
	if newFilePath != nil {
		// This code should be updated to handle file changes using the Files field and FileRepository
		// For compatibility, we could add a new file to the past exam
		// But this would require modifying the UpdatePastExam method signature to include the new file data
		// For now, we'll just log that this should be handled differently
		logger.Warn().Msg("File URL update requested but file handling has changed to use Files field")
	}

	// Update timestamp
	pastExam.UpdatedAt = time.Now()

	err = s.pastExamRepo.UpdatePastExam(ctx, pastExam)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			// Should not happen often after GetByID check, but handle defensively
			return apperrors.ErrPastExamNotFound
		}
		return fmt.Errorf("error updating past exam: %w", err)
	}

	return nil
}

// DeletePastExam deletes a past exam if the user is the original uploader and an instructor
func (s *pastExamServiceImpl) DeletePastExam(ctx context.Context, id int64, userID int64) error {
	// Validate ID
	if id <= 0 {
		return fmt.Errorf("%w: invalid exam ID", apperrors.ErrValidationFailed)
	}

	// Check if the user can delete this exam
	err := s.authService.ValidatePastExamOwnership(ctx, id, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return apperrors.ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	err = s.pastExamRepo.DeletePastExam(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return apperrors.ErrPastExamNotFound
		}
		return fmt.Errorf("error deleting past exam: %w", err)
	}

	return nil
}

// getUserInfo gets user information
func (s *pastExamServiceImpl) getUserInfo(ctx context.Context, userID int64) (*models.User, error) {
	return s.authService.GetUserInfo(ctx, userID)
}

// AddFileToPastExam adds a file to a past exam
func (s *pastExamServiceImpl) AddFileToPastExam(ctx context.Context, pastExamID int64, file *models.File) (int64, error) {
	// Check if the past exam exists
	_, err := s.pastExamRepo.GetPastExamByID(ctx, pastExamID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return 0, apperrors.ErrPastExamNotFound
		}
		return 0, fmt.Errorf("error checking past exam existence: %w", err)
	}

	// Check if user has permission to add files to this past exam
	err = s.authService.ValidatePastExamOwnership(ctx, pastExamID, file.UploadedBy)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return 0, apperrors.ErrPermissionDenied
		}
		return 0, fmt.Errorf("authorization validation error: %w", err)
	}

	// Save file metadata in database
	fileRepo := s.pastExamRepo.GetFileRepository()
	fileID, err := fileRepo.CreateFile(ctx, file)
	if err != nil {
		return 0, fmt.Errorf("error creating file record: %w", err)
	}

	// Associate file with past exam
	err = fileRepo.AddFileToPastExam(ctx, pastExamID, fileID)
	if err != nil {
		return 0, fmt.Errorf("error associating file with past exam: %w", err)
	}

	return fileID, nil
}

// RemoveFileFromPastExam removes a file from a past exam
func (s *pastExamServiceImpl) RemoveFileFromPastExam(ctx context.Context, pastExamID, fileID, userID int64) error {
	// Check if the past exam exists
	_, err := s.pastExamRepo.GetPastExamByID(ctx, pastExamID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return apperrors.ErrPastExamNotFound
		}
		return fmt.Errorf("error checking past exam existence: %w", err)
	}

	// Check if user has permission to remove files from this past exam
	err = s.authService.ValidatePastExamOwnership(ctx, pastExamID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return apperrors.ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	// Get file repository
	fileRepo := s.pastExamRepo.GetFileRepository()

	// Remove association between file and past exam
	err = fileRepo.RemoveFileFromPastExam(ctx, pastExamID, fileID)
	if err != nil {
		return fmt.Errorf("error removing file association: %w", err)
	}

	// Optionally delete the file itself if it's not used elsewhere
	// This could be implemented separately if needed

	return nil
}

// GetPastExamFiles gets all files associated with a past exam
func (s *pastExamServiceImpl) GetPastExamFiles(ctx context.Context, pastExamID int64) ([]*models.File, error) {
	// Check if the past exam exists
	_, err := s.pastExamRepo.GetPastExamByID(ctx, pastExamID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return nil, apperrors.ErrPastExamNotFound
		}
		return nil, fmt.Errorf("error checking past exam existence: %w", err)
	}

	// Get file repository
	fileRepo := s.pastExamRepo.GetFileRepository()

	// Get all files associated with this past exam
	files, err := fileRepo.GetFilesByPastExamID(ctx, pastExamID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving files: %w", err)
	}

	return files, nil
}
