package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// FacultyService defines the interface for faculty-related operations
type FacultyService interface {
	CreateFaculty(ctx context.Context, faculty *models.Faculty) (int64, error)
	GetFacultyByID(ctx context.Context, id int64) (*models.Faculty, error)
	GetAllFaculties(ctx context.Context) ([]*models.Faculty, error)
	UpdateFaculty(ctx context.Context, faculty *models.Faculty) error
	DeleteFaculty(ctx context.Context, id int64) error
}

// facultyServiceImpl implements the FacultyService interface
type facultyServiceImpl struct {
	facultyRepo *repositories.FacultyRepository
}

// NewFacultyService creates a new faculty service instance
func NewFacultyService(facultyRepo *repositories.FacultyRepository) FacultyService {
	return &facultyServiceImpl{
		facultyRepo: facultyRepo,
	}
}

// validateFaculty validates faculty data before database operations
func (s *facultyServiceImpl) validateFaculty(faculty *models.Faculty) error {
	if faculty == nil {
		return fmt.Errorf("%w: faculty is nil", apperrors.ErrValidationFailed)
	}

	// Validate name
	if strings.TrimSpace(faculty.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", apperrors.ErrValidationFailed)
	}

	// Validate faculty code
	if strings.TrimSpace(faculty.Code) == "" {
		return fmt.Errorf("%w: code cannot be empty", apperrors.ErrValidationFailed)
	}

	// Faculty code should be alphanumeric and uppercase
	if !isValidFacultyCode(faculty.Code) {
		return fmt.Errorf("%w: code must be alphanumeric and uppercase", apperrors.ErrValidationFailed)
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
func (s *facultyServiceImpl) CreateFaculty(ctx context.Context, faculty *models.Faculty) (int64, error) {
	// Validate faculty data
	if err := s.validateFaculty(faculty); err != nil {
		return 0, err
	}

	id, err := s.facultyRepo.CreateFaculty(ctx, faculty)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
			return 0, apperrors.ErrFacultyAlreadyExists
		}
		return 0, fmt.Errorf("error creating faculty: %w", err)
	}
	return id, nil
}

// GetFacultyByID retrieves a faculty by ID
func (s *facultyServiceImpl) GetFacultyByID(ctx context.Context, id int64) (*models.Faculty, error) {
	// Validate ID
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid faculty ID", apperrors.ErrValidationFailed)
	}

	faculty, err := s.facultyRepo.GetFacultyByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return nil, apperrors.ErrFacultyNotFound
		}
		return nil, fmt.Errorf("error retrieving faculty: %w", err)
	}
	return faculty, nil
}

// GetAllFaculties retrieves all faculties
func (s *facultyServiceImpl) GetAllFaculties(ctx context.Context) ([]*models.Faculty, error) {
	faculties, err := s.facultyRepo.GetAllFaculties(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving faculties: %w", err)
	}
	return faculties, nil
}

// UpdateFaculty updates an existing faculty
func (s *facultyServiceImpl) UpdateFaculty(ctx context.Context, faculty *models.Faculty) error {
	// Validate faculty data
	if err := s.validateFaculty(faculty); err != nil {
		return err
	}

	// Validate ID
	if faculty.ID <= 0 {
		return fmt.Errorf("%w: invalid faculty ID", apperrors.ErrValidationFailed)
	}

	err := s.facultyRepo.UpdateFaculty(ctx, faculty)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return apperrors.ErrFacultyNotFound
		}
		if errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
			return apperrors.ErrFacultyAlreadyExists
		}
		return fmt.Errorf("error updating faculty: %w", err)
	}
	return nil
}

// DeleteFaculty deletes a faculty by ID
func (s *facultyServiceImpl) DeleteFaculty(ctx context.Context, id int64) error {
	// Validate ID
	if id <= 0 {
		return fmt.Errorf("%w: invalid faculty ID", apperrors.ErrValidationFailed)
	}

	err := s.facultyRepo.DeleteFaculty(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return apperrors.ErrFacultyNotFound
		}
		// If there's a specific repository error for faculty with departments, handle it here
		return fmt.Errorf("error deleting faculty: %w", err)
	}
	return nil
}
