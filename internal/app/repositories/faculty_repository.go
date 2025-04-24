package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// Faculty error types
var (
	ErrFacultyNotFound      = errors.New("faculty not found")
	ErrFacultyAlreadyExists = errors.New("faculty with this name or code already exists")
)

// FacultyRepository handles faculty database operations
type FacultyRepository struct {
	db *pgxpool.Pool
}

// NewFacultyRepository creates a new FacultyRepository
func NewFacultyRepository(db *pgxpool.Pool) *FacultyRepository {
	return &FacultyRepository{
		db: db,
	}
}

// CreateFaculty creates a new faculty
func (r *FacultyRepository) CreateFaculty(ctx context.Context, faculty *models.Faculty) (int64, error) {
	// Check if faculty already exists
	exists, err := r.FacultyExistsByNameOrCode(ctx, faculty.Name, faculty.Code)
	if err != nil {
		return 0, fmt.Errorf("error checking faculty: %w", err)
	}
	if exists {
		return 0, ErrFacultyAlreadyExists
	}

	var id int64
	err = r.db.QueryRow(ctx, `
		INSERT INTO faculties (name, code) 
		VALUES ($1, $2) 
		RETURNING id`,
		faculty.Name, faculty.Code).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating faculty: %w", err)
	}

	return id, nil
}

// GetFacultyByID retrieves a faculty by ID
func (r *FacultyRepository) GetFacultyByID(ctx context.Context, id int64) (*models.Faculty, error) {
	faculty := &models.Faculty{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, code 
		FROM faculties 
		WHERE id = $1`,
		id).Scan(&faculty.ID, &faculty.Name, &faculty.Code)

	if err != nil {
		return nil, ErrFacultyNotFound
	}

	return faculty, nil
}

// GetAllFaculties retrieves all faculties
func (r *FacultyRepository) GetAllFaculties(ctx context.Context) ([]*models.Faculty, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, code 
		FROM faculties 
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("error querying faculties: %w", err)
	}
	defer rows.Close()

	faculties := []*models.Faculty{}
	for rows.Next() {
		faculty := &models.Faculty{}
		if err := rows.Scan(&faculty.ID, &faculty.Name, &faculty.Code); err != nil {
			return nil, fmt.Errorf("error scanning faculty row: %w", err)
		}
		faculties = append(faculties, faculty)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating faculty rows: %w", err)
	}

	return faculties, nil
}

// FacultyExistsByNameOrCode checks if a faculty exists by name or code
func (r *FacultyRepository) FacultyExistsByNameOrCode(ctx context.Context, name, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM faculties WHERE name = $1 OR code = $2)`,
		name, code).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking faculty existence: %w", err)
	}

	return exists, nil
}

// UpdateFaculty updates an existing faculty
func (r *FacultyRepository) UpdateFaculty(ctx context.Context, faculty *models.Faculty) error {
	// Check if faculty exists
	_, err := r.GetFacultyByID(ctx, faculty.ID)
	if err != nil {
		return ErrFacultyNotFound
	}

	// Check if faculty name or code is already used by another faculty
	var exists bool
	err = r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM faculties WHERE (name = $1 OR code = $2) AND id != $3)`,
		faculty.Name, faculty.Code, faculty.ID).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking faculty uniqueness: %w", err)
	}

	if exists {
		return ErrFacultyAlreadyExists
	}

	// Update the faculty
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE faculties 
		SET name = $1, code = $2 
		WHERE id = $3`,
		faculty.Name, faculty.Code, faculty.ID)

	if err != nil {
		return fmt.Errorf("error updating faculty: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrFacultyNotFound
	}

	return nil
}

// DeleteFaculty deletes a faculty by ID
func (r *FacultyRepository) DeleteFaculty(ctx context.Context, id int64) error {
	// Check if faculty exists
	_, err := r.GetFacultyByID(ctx, id)
	if err != nil {
		return ErrFacultyNotFound
	}

	// Check if faculty has associated departments
	var hasDepartments bool
	err = r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM departments WHERE faculty_id = $1)`,
		id).Scan(&hasDepartments)

	if err != nil {
		return fmt.Errorf("error checking associated departments: %w", err)
	}

	if hasDepartments {
		return errors.New("faculty has associated departments and cannot be deleted")
	}

	// Delete the faculty
	cmdTag, err := r.db.Exec(ctx, `
		DELETE FROM faculties 
		WHERE id = $1`,
		id)

	if err != nil {
		return fmt.Errorf("error deleting faculty: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrFacultyNotFound
	}

	return nil
}
