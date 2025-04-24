package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

var (
	ErrStudentIDExists = errors.New("student ID already in use")
)

// StudentRepository handles student database operations
type StudentRepository struct {
	db *pgxpool.Pool
}

// NewStudentRepository creates a new StudentRepository
func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{
		db: db,
	}
}

// CreateStudent creates a new student
func (r *StudentRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	// Check student ID availability
	exists, err := r.StudentIDExists(ctx, student.StudentID)
	if err != nil {
		return fmt.Errorf("error checking student ID: %w", err)
	}
	if exists {
		return ErrStudentIDExists
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO students (user_id, student_id, department_id, graduation_year)
		VALUES ($1, $2, $3, $4)`,
		student.UserID, student.StudentID, student.DepartmentID, student.GraduationYear)

	if err != nil {
		return fmt.Errorf("error creating student: %w", err)
	}

	return nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *StudentRepository) GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error) {
	student := &models.Student{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, student_id, department_id, graduation_year
		FROM students
		WHERE user_id = $1`,
		userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.DepartmentID, &student.GraduationYear)

	if err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	return student, nil
}

// StudentIDExists checks if a student ID already exists
func (r *StudentRepository) StudentIDExists(ctx context.Context, studentID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM students WHERE student_id = $1)`,
		studentID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking student ID: %w", err)
	}

	return exists, nil
}
 