package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
)

// Common past exam errors
var (
	ErrPastExamNotFound = errors.New("past exam not found")
	ErrPermissionDenied = errors.New("you don't have permission for this action")
	ErrInstructorOnly   = errors.New("only instructors can create past exams")
	ErrValidationFailed = errors.New("validation failed")
)

// PastExamService handles past exam related operations
type PastExamService struct {
	pastExamRepo *repositories.PastExamRepository
	authService  *auth.AuthorizationService
}

// NewPastExamService creates a new past exam service instance
func NewPastExamService(pastExamRepo *repositories.PastExamRepository, authService *auth.AuthorizationService) *PastExamService {
	return &PastExamService{
		pastExamRepo: pastExamRepo,
		authService:  authService,
	}
}

// validatePastExam validates the past exam data before database operations
func (s *PastExamService) validatePastExam(pastExam *models.PastExam) error {
	if pastExam == nil {
		return fmt.Errorf("%w: past exam is nil", ErrValidationFailed)
	}

	// Validate year
	currentYear := time.Now().Year()
	if pastExam.Year < 1900 || pastExam.Year > currentYear+1 {
		return fmt.Errorf("%w: year must be between 1900 and %d", ErrValidationFailed, currentYear+1)
	}

	// Validate term
	term := strings.ToUpper(string(pastExam.Term))
	if term != "FALL" && term != "SPRING" {
		return fmt.Errorf("%w: term must be FALL or SPRING", ErrValidationFailed)
	}

	// Validate department ID
	if pastExam.DepartmentID <= 0 {
		return fmt.Errorf("%w: department ID must be positive", ErrValidationFailed)
	}

	// Validate course code
	if strings.TrimSpace(pastExam.CourseCode) == "" {
		return fmt.Errorf("%w: course code cannot be empty", ErrValidationFailed)
	}

	// Validate title
	if strings.TrimSpace(pastExam.Title) == "" {
		return fmt.Errorf("%w: title cannot be empty", ErrValidationFailed)
	}

	// Validate content
	if strings.TrimSpace(pastExam.Content) == "" {
		return fmt.Errorf("%w: content cannot be empty", ErrValidationFailed)
	}

	// Validate file URL if provided
	if pastExam.FileURL != "" {
		_, err := url.ParseRequestURI(pastExam.FileURL)
		if err != nil {
			return fmt.Errorf("%w: invalid file URL: %v", ErrValidationFailed, err)
		}
	}

	return nil
}

// CreatePastExam creates a new past exam, only accessible to instructors
func (s *PastExamService) CreatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64) (int64, error) {
	// Validate past exam data
	if err := s.validatePastExam(pastExam); err != nil {
		return 0, err
	}

	// Validate that the user is an instructor and get instructor ID
	instructor, err := s.authService.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrNotInstructor) {
			return 0, ErrInstructorOnly
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
func (s *PastExamService) GetPastExamByID(ctx context.Context, id int64) (*models.PastExam, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid ID", ErrValidationFailed)
	}

	pastExam, err := s.pastExamRepo.GetPastExamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return nil, ErrPastExamNotFound
		}
		return nil, fmt.Errorf("error retrieving past exam: %w", err)
	}

	return pastExam, nil
}

// GetAllPastExams retrieves all past exams with pagination and filtering
func (s *PastExamService) GetAllPastExams(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]models.PastExam, int, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100 // Maximum page size to prevent excessive loads
	}

	// Validate filter values if present
	if filters != nil {
		// Year validation
		if yearVal, ok := filters["year"]; ok && yearVal != nil {
			if year, ok := yearVal.(int); ok {
				currentYear := time.Now().Year()
				if year < 1900 || year > currentYear+1 {
					return nil, 0, fmt.Errorf("%w: year must be between 1900 and %d", ErrValidationFailed, currentYear+1)
				}
			}
		}

		// Term validation
		if termVal, ok := filters["term"]; ok && termVal != nil {
			if term, ok := termVal.(string); ok {
				termUpper := strings.ToUpper(term)
				if termUpper != "FALL" && termUpper != "SPRING" {
					return nil, 0, fmt.Errorf("%w: term must be FALL or SPRING", ErrValidationFailed)
				}
				filters["term"] = termUpper // Normalize the term to uppercase
			}
		}
	}

	pastExams, totalCount, err := s.pastExamRepo.GetAllPastExams(ctx, page, pageSize, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving past exams: %w", err)
	}

	return pastExams, totalCount, nil
}

// UpdatePastExam updates an existing past exam if the user is the original uploader and an instructor
func (s *PastExamService) UpdatePastExam(ctx context.Context, pastExam *models.PastExam, userID int64) error {
	// Validate past exam data
	if err := s.validatePastExam(pastExam); err != nil {
		return err
	}

	// Validate exam ID
	if pastExam.ID <= 0 {
		return fmt.Errorf("%w: invalid exam ID", ErrValidationFailed)
	}

	// Check if the exam exists
	existingExam, err := s.pastExamRepo.GetPastExamByID(ctx, pastExam.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return ErrPastExamNotFound
		}
		return fmt.Errorf("error checking exam existence: %w", err)
	}

	// Check if the user can modify this exam
	err = s.authService.ValidatePastExamOwnership(ctx, pastExam.ID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	// Preserve original instructor information
	pastExam.InstructorID = existingExam.InstructorID
	
	// Preserve display information
	pastExam.UploadedByName = existingExam.UploadedByName
	pastExam.UploadedByEmail = existingExam.UploadedByEmail

	// Update timestamp
	pastExam.UpdatedAt = time.Now()

	err = s.pastExamRepo.UpdatePastExam(ctx, pastExam)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return ErrPastExamNotFound
		}
		return fmt.Errorf("error updating past exam: %w", err)
	}

	return nil
}

// DeletePastExam deletes a past exam if the user is the original uploader and an instructor
func (s *PastExamService) DeletePastExam(ctx context.Context, id int64, userID int64) error {
	// Validate ID
	if id <= 0 {
		return fmt.Errorf("%w: invalid exam ID", ErrValidationFailed)
	}

	// Check if the user can delete this exam
	err := s.authService.ValidatePastExamOwnership(ctx, id, userID)
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("authorization validation error: %w", err)
	}

	err = s.pastExamRepo.DeletePastExam(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return ErrPastExamNotFound
		}
		return fmt.Errorf("error deleting past exam: %w", err)
	}

	return nil
}

// getUserInfo gets user information
func (s *PastExamService) getUserInfo(ctx context.Context, userID int64) (*models.User, error) {
	return s.authService.GetUserInfo(ctx, userID)
}
