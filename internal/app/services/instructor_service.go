package services

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// InstructorService defines the interface for instructor-related operations
type InstructorService interface {
	GetInstructorByID(ctx context.Context, instructorID int64) (*models.Instructor, error)
	GetInstructorsByDepartment(ctx context.Context, departmentID int64) ([]*models.Instructor, error)
	GetInstructorWithDetails(ctx context.Context, userID int64) (*models.Instructor, error)
	UpdateTitle(ctx context.Context, userID int64, newTitle string) error
}

// instructorServiceImpl implements the InstructorService interface
type instructorServiceImpl struct {
	userRepo       *repositories.UserRepository
	departmentRepo *repositories.DepartmentRepository
}

// NewInstructorService creates a new instructor service instance
func NewInstructorService(userRepo *repositories.UserRepository, departmentRepo *repositories.DepartmentRepository) InstructorService {
	return &instructorServiceImpl{
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
	}
}

// validateUserID validates a user ID
func (s *instructorServiceImpl) validateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", apperrors.ErrValidationFailed)
	}
	return nil
}

// validateTitle validates an instructor's title
func (s *instructorServiceImpl) validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%w: title cannot be empty", apperrors.ErrValidationFailed)
	}

	// Just check that title is reasonable length and doesn't contain invalid characters
	if len(title) > 100 {
		return fmt.Errorf("%w: title is too long (max 100 characters)", apperrors.ErrValidationFailed)
	}

	// Check that title contains only allowed characters for academic titles
	// (letters, spaces, dots, and hyphens)
	for _, char := range title {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) && char != '.' && char != '-' {
			return fmt.Errorf("%w: title contains invalid characters", apperrors.ErrValidationFailed)
		}
	}

	return nil
}

// validateDepartmentID validates a department ID
func (s *instructorServiceImpl) validateDepartmentID(ctx context.Context, departmentID int64) error {
	if departmentID <= 0 {
		return fmt.Errorf("%w: department ID must be positive", apperrors.ErrValidationFailed)
	}

	// Check if department exists
	department, err := s.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return fmt.Errorf("error verifying department existence: %w", err)
	}

	if department == nil {
		return apperrors.ErrDepartmentNotFound
	}

	return nil
}

// GetInstructorByID retrieves an instructor by user ID with department details
func (s *instructorServiceImpl) GetInstructorByID(ctx context.Context, userID int64) (*models.Instructor, error) {
	// Validate the user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get the instructor information
	instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting instructor: %w", err)
	}

	// Get the user information
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Attach user to instructor
	instructor.User = user

	return instructor, nil
}

// GetInstructorWithDetails retrieves an instructor with all details including department
func (s *instructorServiceImpl) GetInstructorWithDetails(ctx context.Context, userID int64) (*models.Instructor, error) {
	// Validate the user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get instructor with basic user info
	instructor, err := s.GetInstructorByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get department details from the user model, if available
	if instructor.User != nil && instructor.User.DepartmentID != nil {
		departmentID := *instructor.User.DepartmentID

		// Get department details
		department, err := s.departmentRepo.GetByID(ctx, departmentID)
		if err == nil && department != nil {
			// Attach department to instructor
			instructor.Department = department
		}
	}

	return instructor, nil
}

// GetInstructorsByDepartment retrieves all instructors in a department
func (s *instructorServiceImpl) GetInstructorsByDepartment(ctx context.Context, departmentID int64) ([]*models.Instructor, error) {
	// This method is now deprecated. The implementation is kept to avoid breaking changes
	// but returns an empty slice since the functionality has been moved to UserService.
	return []*models.Instructor{}, nil
}

// UpdateTitle updates an instructor's title
func (s *instructorServiceImpl) UpdateTitle(ctx context.Context, userID int64, newTitle string) error {
	// This method is now deprecated
	// Return not implemented error
	return fmt.Errorf("method deprecated")
}
