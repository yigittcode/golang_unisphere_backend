package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// Department error types
var (
	ErrDepartmentNotFound      = errors.New("department not found")
	ErrDepartmentAlreadyExists = errors.New("department with this name or code already exists")
)

// DepartmentRepository handles database operations for departments
type DepartmentRepository struct {
	db *pgxpool.Pool
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{
		db: db,
	}
}

// Create creates a new department
func (r *DepartmentRepository) Create(ctx context.Context, department *models.Department) error {
	query := `
		INSERT INTO departments (faculty_id, name, code)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, department.FacultyID, department.Name, department.Code).Scan(&department.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a department by ID
func (r *DepartmentRepository) GetByID(ctx context.Context, id int64) (*models.Department, error) {
	query := `
		SELECT id, faculty_id, name, code
		FROM departments
		WHERE id = $1
	`

	var department models.Department
	err := r.db.QueryRow(ctx, query, id).Scan(
		&department.ID,
		&department.FacultyID,
		&department.Name,
		&department.Code,
	)

	if err != nil {
		return nil, fmt.Errorf("error retrieving department: %w", err)
	}

	return &department, nil
}

// GetAll retrieves all departments
func (r *DepartmentRepository) GetAll(ctx context.Context) ([]*models.Department, error) {
	query := `
		SELECT id, faculty_id, name, code
		FROM departments
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		var department models.Department
		if err := rows.Scan(
			&department.ID,
			&department.FacultyID,
			&department.Name,
			&department.Code,
		); err != nil {
			return nil, err
		}
		departments = append(departments, &department)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return departments, nil
}

// GetByFacultyID retrieves all departments for a given faculty
func (r *DepartmentRepository) GetByFacultyID(ctx context.Context, facultyID int64) ([]*models.Department, error) {
	query := `
		SELECT id, faculty_id, name, code
		FROM departments
		WHERE faculty_id = $1
	`

	rows, err := r.db.Query(ctx, query, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		var department models.Department
		if err := rows.Scan(
			&department.ID,
			&department.FacultyID,
			&department.Name,
			&department.Code,
		); err != nil {
			return nil, err
		}
		departments = append(departments, &department)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return departments, nil
}

// DepartmentExistsByNameOrCode checks if a department exists by name or code
func (r *DepartmentRepository) DepartmentExistsByNameOrCode(ctx context.Context, name, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM departments WHERE name = $1 OR code = $2)`,
		name, code).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking department existence: %w", err)
	}

	return exists, nil
}

// Update updates an existing department
func (r *DepartmentRepository) Update(ctx context.Context, department *models.Department) error {
	// Check if department exists
	existingDept, err := r.GetByID(ctx, department.ID)
	if err != nil {
		return ErrDepartmentNotFound
	}

	if existingDept == nil {
		return ErrDepartmentNotFound
	}

	// Check if department name or code is already used by another department
	var exists bool
	err = r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM departments WHERE (name = $1 OR code = $2) AND id != $3)`,
		department.Name, department.Code, department.ID).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking department uniqueness: %w", err)
	}

	if exists {
		return ErrDepartmentAlreadyExists
	}

	// Update the department
	query := `
		UPDATE departments 
		SET faculty_id = $1, name = $2, code = $3 
		WHERE id = $4
	`

	cmdTag, err := r.db.Exec(ctx, query,
		department.FacultyID, department.Name, department.Code, department.ID)

	if err != nil {
		return fmt.Errorf("error updating department: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}

	return nil
}

// Delete deletes a department by ID
func (r *DepartmentRepository) Delete(ctx context.Context, id int64) error {
	// Check if department exists
	existingDept, err := r.GetByID(ctx, id)
	if err != nil {
		return ErrDepartmentNotFound
	}

	if existingDept == nil {
		return ErrDepartmentNotFound
	}

	// Check if department has associated entities (like courses, instructors, students)
	// This would depend on your schema and relationships
	// For example, check for past exams in this department
	var hasRelatedEntities bool
	err = r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM past_exams WHERE department_id = $1)`,
		id).Scan(&hasRelatedEntities)

	if err != nil {
		return fmt.Errorf("error checking related entities: %w", err)
	}

	if hasRelatedEntities {
		return errors.New("department has associated data and cannot be deleted")
	}

	// Delete the department
	query := `DELETE FROM departments WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("error deleting department: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}

	return nil
}
