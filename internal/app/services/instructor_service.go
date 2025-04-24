package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
)

// Common instructor errors
var (
	ErrInstructorNotFound     = errors.New("instructor not found")
	ErrUnauthorized           = errors.New("user is not an instructor")
	ErrInstructorValidation   = errors.New("instructor validation failed")
	ErrDepartmentNotAvailable = errors.New("department not available")
)

// InstructorService handles instructor-related operations
type InstructorService struct {
	userRepo       *repositories.UserRepository
	departmentRepo *repositories.DepartmentRepository
}

// NewInstructorService creates a new instructor service instance
func NewInstructorService(userRepo *repositories.UserRepository, departmentRepo *repositories.DepartmentRepository) *InstructorService {
	return &InstructorService{
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
	}
}

// validateUserID validates a user ID
func (s *InstructorService) validateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", ErrInstructorValidation)
	}
	return nil
}

// validateTitle validates an instructor's title
func (s *InstructorService) validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("%w: title cannot be empty", ErrInstructorValidation)
	}

	// Just check that title is reasonable length and doesn't contain invalid characters
	if len(title) > 100 {
		return fmt.Errorf("%w: title is too long (max 100 characters)", ErrInstructorValidation)
	}

	// Check that title contains only allowed characters for academic titles
	// (letters, spaces, dots, and hyphens)
	for _, char := range title {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) && char != '.' && char != '-' {
			return fmt.Errorf("%w: title contains invalid characters", ErrInstructorValidation)
		}
	}

	return nil
}

// validateDepartmentID validates a department ID
func (s *InstructorService) validateDepartmentID(ctx context.Context, departmentID int64) error {
	if departmentID <= 0 {
		return fmt.Errorf("%w: department ID must be positive", ErrInstructorValidation)
	}

	// Check if department exists
	department, err := s.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return fmt.Errorf("error verifying department existence: %w", err)
	}

	if department == nil {
		return ErrDepartmentNotAvailable
	}

	return nil
}

// GetInstructorByID retrieves an instructor by user ID with department details
func (s *InstructorService) GetInstructorByID(ctx context.Context, userID int64) (*models.Instructor, error) {
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
func (s *InstructorService) GetInstructorWithDetails(ctx context.Context, userID int64) (*models.Instructor, error) {
	// Validate the user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get instructor with basic user info
	instructor, err := s.GetInstructorByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get department details
	department, err := s.departmentRepo.GetByID(ctx, instructor.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("error getting department details: %w", err)
	}

	// Attach department to instructor
	instructor.Department = department

	return instructor, nil
}

// GetInstructorsByDepartment retrieves all instructors in a department
func (s *InstructorService) GetInstructorsByDepartment(ctx context.Context, departmentID int64) ([]*models.Instructor, error) {
	// Validate department ID
	if err := s.validateDepartmentID(ctx, departmentID); err != nil {
		return nil, err
	}

	// Check if department exists
	department, err := s.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("error getting department: %w", err)
	}
	if department == nil {
		return nil, fmt.Errorf("department not found")
	}

	// Get instructors by department ID
	instructors, err := s.userRepo.GetInstructorsByDepartmentID(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("error getting instructors: %w", err)
	}

	// Add department to each instructor
	for _, instructor := range instructors {
		instructor.Department = department
	}

	return instructors, nil
}

// UpdateInstructorTitle updates an instructor's title
func (s *InstructorService) UpdateInstructorTitle(ctx context.Context, userID int64, newTitle string) error {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return err
	}

	// Validate title
	if err := s.validateTitle(newTitle); err != nil {
		return err
	}

	// Check if the instructor exists
	instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting instructor: %w", err)
	}
	if instructor == nil {
		return ErrInstructorNotFound
	}

	// Check if the user is an instructor
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if user.RoleType != models.RoleInstructor {
		return ErrUnauthorized
	}

	// Update the instructor's title
	if instructor.Title == newTitle {
		return nil // Title already set to the requested value
	}

	// Call the repository method to update the title
	return s.userRepo.UpdateInstructorTitle(ctx, userID, newTitle)
}
