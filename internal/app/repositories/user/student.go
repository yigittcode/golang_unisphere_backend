package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/dberrors"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

var (
	ErrStudentIDExists = errors.New("student ID already in use")
	ErrStudentNotFound = ErrUserNotFound
)

// StudentRepository handles student database operations
type StudentRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewStudentRepository creates a new StudentRepository
func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// CreateStudent creates a new student
func (r *StudentRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	sql, args, err := r.sb.Insert("students").
		Columns("user_id", "student_id", "department_id", "graduation_year").
		Values(student.UserID, student.StudentID, student.DepartmentID, student.GraduationYear).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create student SQL")
		return fmt.Errorf("failed to build create student query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		if dberrors.IsDuplicateConstraintError(err, "students_student_id_key") {
			logger.Warn().Str("studentID", student.StudentID).Msg("Attempted to create student with duplicate student ID")
			return ErrStudentIDExists
		}
		logger.Error().Err(err).Int64("userID", student.UserID).Str("studentID", student.StudentID).Msg("Error executing create student query")
		return fmt.Errorf("error creating student: %w", err)
	}

	logger.Info().Int64("userID", student.UserID).Str("studentID", student.StudentID).Msg("Student created successfully")
	return nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *StudentRepository) GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error) {
	var student models.Student
	sql, args, err := r.sb.Select("id", "user_id", "student_id", "department_id", "graduation_year").
		From("students").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get student by user ID SQL")
		return nil, fmt.Errorf("failed to build get student query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.DepartmentID, &student.GraduationYear)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Int64("userID", userID).Msg("Student not found by user ID")
			return nil, ErrStudentNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error scanning student row")
		return nil, fmt.Errorf("error retrieving student: %w", err)
	}

	return &student, nil
}

// StudentIDExists checks if a student ID already exists
func (r *StudentRepository) StudentIDExists(ctx context.Context, studentID string) (bool, error) {
	var exists bool
	sql, args, err := r.sb.Select("1").
		From("students").
		Where(squirrel.Eq{"student_id": studentID}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building student ID exists SQL")
		return false, fmt.Errorf("failed to build student ID exists query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Str("studentID", studentID).Msg("Error checking student ID existence")
		return false, fmt.Errorf("error checking student ID existence: %w", err)
	}

	return exists, nil
}
