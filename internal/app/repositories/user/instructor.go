package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// InstructorRepository handles instructor database operations
type InstructorRepository struct {
	db *pgxpool.Pool
}

// NewInstructorRepository creates a new InstructorRepository
func NewInstructorRepository(db *pgxpool.Pool) *InstructorRepository {
	return &InstructorRepository{
		db: db,
	}
}

// CreateInstructor creates a new instructor
func (r *InstructorRepository) CreateInstructor(ctx context.Context, instructor *models.Instructor) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO instructors (user_id, department_id, title)
		VALUES ($1, $2, $3)`,
		instructor.UserID, instructor.DepartmentID, instructor.Title)

	if err != nil {
		return fmt.Errorf("error creating instructor: %w", err)
	}

	return nil
}

// GetInstructorByUserID retrieves an instructor by user ID
func (r *InstructorRepository) GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error) {
	instructor := &models.Instructor{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, department_id, title
		FROM instructors
		WHERE user_id = $1`,
		userID).Scan(
		&instructor.ID, &instructor.UserID, &instructor.DepartmentID, &instructor.Title)

	if err != nil {
		return nil, fmt.Errorf("instructor not found: %w", err)
	}

	return instructor, nil
}

// GetInstructorsByDepartmentID retrieves all instructors in a department
func (r *InstructorRepository) GetInstructorsByDepartmentID(ctx context.Context, departmentID int64) ([]*models.Instructor, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, i.user_id, i.department_id, i.title, 
		       u.first_name, u.last_name, u.email 
		FROM instructors i
		JOIN users u ON i.user_id = u.id
		WHERE i.department_id = $1
		ORDER BY u.last_name, u.first_name`,
		departmentID)
	if err != nil {
		return nil, fmt.Errorf("error getting instructors: %w", err)
	}
	defer rows.Close()

	var instructors []*models.Instructor
	for rows.Next() {
		instructor := &models.Instructor{
			User: &models.User{},
		}
		err := rows.Scan(
			&instructor.ID,
			&instructor.UserID,
			&instructor.DepartmentID,
			&instructor.Title,
			&instructor.User.FirstName,
			&instructor.User.LastName,
			&instructor.User.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning instructor: %w", err)
		}
		instructors = append(instructors, instructor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating instructors: %w", err)
	}

	return instructors, nil
}

// UpdateInstructorTitle updates an instructor's title
func (r *InstructorRepository) UpdateInstructorTitle(ctx context.Context, userID int64, newTitle string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE instructors
		SET title = $1
		WHERE user_id = $2`,
		newTitle, userID)

	if err != nil {
		return fmt.Errorf("error updating instructor title: %w", err)
	}

	return nil
}
 