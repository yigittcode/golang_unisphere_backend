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

// Move this error to apperrors package in the future
var (
	ErrFacultyForDeptNotFound = errors.New("faculty for department not found")
)

// DepartmentService defines the interface for department-related operations
type DepartmentService interface {
	CreateDepartment(ctx context.Context, department *models.Department) error
	GetDepartmentByID(ctx context.Context, id int64) (*models.Department, error)
	GetAllDepartments(ctx context.Context) ([]*models.Department, error)
	GetDepartmentsByFacultyID(ctx context.Context, facultyID int64) ([]*models.Department, error)
	UpdateDepartment(ctx context.Context, department *models.Department) error
	DeleteDepartment(ctx context.Context, id int64) error
}

// departmentServiceImpl implements the DepartmentService interface
type departmentServiceImpl struct {
	departmentRepo *repositories.DepartmentRepository
	facultyRepo    *repositories.FacultyRepository
}

// NewDepartmentService creates a new department service instance
func NewDepartmentService(departmentRepo *repositories.DepartmentRepository, facultyRepo *repositories.FacultyRepository) DepartmentService {
	return &departmentServiceImpl{
		departmentRepo: departmentRepo,
		facultyRepo:    facultyRepo,
	}
}

// validateDepartment validates department data before database operations
func (s *departmentServiceImpl) validateDepartment(department *models.Department) error {
	if department == nil {
		return fmt.Errorf("%w: department is nil", apperrors.ErrValidationFailed)
	}

	// Validate faculty ID
	if department.FacultyID <= 0 {
		return fmt.Errorf("%w: faculty ID must be positive", apperrors.ErrValidationFailed)
	}

	// Validate name
	if strings.TrimSpace(department.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", apperrors.ErrValidationFailed)
	}

	// Validate code
	if strings.TrimSpace(department.Code) == "" {
		return fmt.Errorf("%w: code cannot be empty", apperrors.ErrValidationFailed)
	}

	// Department code should be alphanumeric and uppercase
	if !isValidDepartmentCode(department.Code) {
		return fmt.Errorf("%w: code must be alphanumeric and uppercase", apperrors.ErrValidationFailed)
	}

	return nil
}

// isValidDepartmentCode checks if a department code is valid
func isValidDepartmentCode(code string) bool {
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

// CreateDepartment creates a new department
func (s *departmentServiceImpl) CreateDepartment(ctx context.Context, department *models.Department) error {
	// Validate department data
	if err := s.validateDepartment(department); err != nil {
		return err
	}

	// First validate that the faculty exists
	faculty, err := s.facultyRepo.GetFacultyByID(ctx, department.FacultyID)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return ErrFacultyForDeptNotFound
		}
		return fmt.Errorf("error checking faculty: %w", err)
	}

	if faculty == nil {
		return ErrFacultyForDeptNotFound
	}

	err = s.departmentRepo.Create(ctx, department)
	if err != nil {
		if errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			return apperrors.ErrDepartmentAlreadyExists
		}
		return fmt.Errorf("error creating department: %w", err)
	}
	return nil
}

// GetDepartmentByID retrieves a department by ID
func (s *departmentServiceImpl) GetDepartmentByID(ctx context.Context, id int64) (*models.Department, error) {
	// Validate ID
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid department ID", apperrors.ErrValidationFailed)
	}

	department, err := s.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving department: %w", err)
	}

	if department == nil {
		return nil, apperrors.ErrDepartmentNotFound
	}

	// Get faculty details and attach to department
	faculty, err := s.facultyRepo.GetFacultyByID(ctx, department.FacultyID)
	if err == nil && faculty != nil {
		department.Faculty = faculty
	}

	return department, nil
}

// GetAllDepartments retrieves all departments
func (s *departmentServiceImpl) GetAllDepartments(ctx context.Context) ([]*models.Department, error) {
	departments, err := s.departmentRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving departments: %w", err)
	}

	// Enrich departments with faculty information
	for _, department := range departments {
		faculty, err := s.facultyRepo.GetFacultyByID(ctx, department.FacultyID)
		if err == nil && faculty != nil {
			department.Faculty = faculty
		}
	}

	return departments, nil
}

// GetDepartmentsByFacultyID retrieves all departments for a specific faculty
func (s *departmentServiceImpl) GetDepartmentsByFacultyID(ctx context.Context, facultyID int64) ([]*models.Department, error) {
	// Validate faculty ID
	if facultyID <= 0 {
		return nil, fmt.Errorf("%w: invalid faculty ID", apperrors.ErrValidationFailed)
	}

	// First check if faculty exists
	faculty, err := s.facultyRepo.GetFacultyByID(ctx, facultyID)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return nil, apperrors.ErrFacultyNotFound
		}
		return nil, fmt.Errorf("error checking faculty: %w", err)
	}

	if faculty == nil {
		return nil, apperrors.ErrFacultyNotFound
	}

	departments, err := s.departmentRepo.GetByFacultyID(ctx, facultyID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving departments by faculty: %w", err)
	}

	// Set faculty for each department
	for _, department := range departments {
		department.Faculty = faculty
	}

	return departments, nil
}

// UpdateDepartment updates an existing department
func (s *departmentServiceImpl) UpdateDepartment(ctx context.Context, department *models.Department) error {
	// Validate department data
	if err := s.validateDepartment(department); err != nil {
		return err
	}

	// Validate ID
	if department.ID <= 0 {
		return fmt.Errorf("%w: invalid department ID", apperrors.ErrValidationFailed)
	}

	// First validate that the faculty exists if faculty ID is changed
	faculty, err := s.facultyRepo.GetFacultyByID(ctx, department.FacultyID)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			return ErrFacultyForDeptNotFound
		}
		return fmt.Errorf("error checking faculty: %w", err)
	}

	if faculty == nil {
		return ErrFacultyForDeptNotFound
	}

	err = s.departmentRepo.Update(ctx, department)
	if err != nil {
		if errors.Is(err, apperrors.ErrDepartmentNotFound) {
			return apperrors.ErrDepartmentNotFound
		}
		if errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			return apperrors.ErrDepartmentAlreadyExists
		}
		return fmt.Errorf("error updating department: %w", err)
	}
	return nil
}

// DeleteDepartment deletes a department by ID
func (s *departmentServiceImpl) DeleteDepartment(ctx context.Context, id int64) error {
	// Validate ID
	if id <= 0 {
		return fmt.Errorf("%w: invalid department ID", apperrors.ErrValidationFailed)
	}

	err := s.departmentRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrDepartmentNotFound) {
			return apperrors.ErrDepartmentNotFound
		}
		// If there's a specific repository error for department with references, handle it here
		return fmt.Errorf("error deleting department: %w", err)
	}
	return nil
}
