package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
)

// Common faculty errors
var (
	ErrFacultyNotFound      = errors.New("faculty not found")
	ErrFacultyAlreadyExists = errors.New("faculty with this name or code already exists")
	ErrFacultyValidation    = errors.New("faculty validation failed")
)

// FacultyService handles faculty-related operations
type FacultyService struct {
	facultyRepo *repositories.FacultyRepository
}

// NewFacultyService creates a new faculty service instance
func NewFacultyService(facultyRepo *repositories.FacultyRepository) *FacultyService {
	return &FacultyService{
		facultyRepo: facultyRepo,
	}
}

// validateFaculty validates faculty data before database operations
func (s *FacultyService) validateFaculty(faculty *models.Faculty) error {
	if faculty == nil {
		return fmt.Errorf("%w: faculty is nil", ErrFacultyValidation)
	}

	// Validate name
	if strings.TrimSpace(faculty.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrFacultyValidation)
	}

	// Validate faculty code
	if strings.TrimSpace(faculty.Code) == "" {
		return fmt.Errorf("%w: code cannot be empty", ErrFacultyValidation)
	}

	// Faculty code should be alphanumeric and uppercase
	if !isValidFacultyCode(faculty.Code) {
		return fmt.Errorf("%w: code must be alphanumeric and uppercase", ErrFacultyValidation)
	}

	return nil
}

// isValidFacultyCode checks if a faculty code is valid
func isValidFacultyCode(code string) bool {
	// Code should be uppercase alphanumeric
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}

	// Ensure code is uppercase
	if code != strings.ToUpper(code) {
		return false
	}

	// Check if code contains only letters and numbers
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// CreateFaculty creates a new faculty
func (s *FacultyService) CreateFaculty(ctx context.Context, faculty *models.Faculty) (int64, error) {
	// Validate faculty data
	if err := s.validateFaculty(faculty); err != nil {
		return 0, err
	}

	id, err := s.facultyRepo.CreateFaculty(ctx, faculty)
	if err != nil {
		if errors.Is(err, repositories.ErrFacultyAlreadyExists) {
			return 0, ErrFacultyAlreadyExists
		}
		return 0, fmt.Errorf("error creating faculty: %w", err)
	}
	return id, nil
}

// GetFacultyByID retrieves a faculty by ID
func (s *FacultyService) GetFacultyByID(ctx context.Context, id int64) (*models.Faculty, error) {
	// Validate ID
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid faculty ID", ErrFacultyValidation)
	}

	faculty, err := s.facultyRepo.GetFacultyByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrFacultyNotFound) {
			return nil, ErrFacultyNotFound
		}
		return nil, fmt.Errorf("error retrieving faculty: %w", err)
	}
	return faculty, nil
}

// GetAllFaculties retrieves all faculties
func (s *FacultyService) GetAllFaculties(ctx context.Context) ([]*models.Faculty, error) {
	faculties, err := s.facultyRepo.GetAllFaculties(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving faculties: %w", err)
	}
	return faculties, nil
}

// UpdateFaculty updates an existing faculty
func (s *FacultyService) UpdateFaculty(ctx context.Context, faculty *models.Faculty) error {
	// Validate faculty data
	if err := s.validateFaculty(faculty); err != nil {
		return err
	}

	// Validate ID
	if faculty.ID <= 0 {
		return fmt.Errorf("%w: invalid faculty ID", ErrFacultyValidation)
	}

	err := s.facultyRepo.UpdateFaculty(ctx, faculty)
	if err != nil {
		if errors.Is(err, repositories.ErrFacultyNotFound) {
			return ErrFacultyNotFound
		}
		if errors.Is(err, repositories.ErrFacultyAlreadyExists) {
			return ErrFacultyAlreadyExists
		}
		return fmt.Errorf("error updating faculty: %w", err)
	}
	return nil
}

// DeleteFaculty deletes a faculty by ID
func (s *FacultyService) DeleteFaculty(ctx context.Context, id int64) error {
	// Validate ID
	if id <= 0 {
		return fmt.Errorf("%w: invalid faculty ID", ErrFacultyValidation)
	}

	err := s.facultyRepo.DeleteFaculty(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrFacultyNotFound) {
			return ErrFacultyNotFound
		}
		// If there's a specific repository error for faculty with departments, handle it here
		return fmt.Errorf("error deleting faculty: %w", err)
	}
	return nil
}
