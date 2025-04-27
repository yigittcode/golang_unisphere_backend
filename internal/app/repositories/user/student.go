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
	ErrIdentifierExists = errors.New("student identifier already in use")
	ErrStudentNotFound  = ErrUserNotFound
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
		Columns("user_id", "identifier", "department_id", "graduation_year").
		Values(student.UserID, student.Identifier, student.DepartmentID, student.GraduationYear).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create student SQL")
		return fmt.Errorf("failed to build create student query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		if dberrors.IsDuplicateConstraintError(err, "students_identifier_key") {
			logger.Warn().Str("identifier", student.Identifier).Msg("Attempted to create student with duplicate identifier")
			return ErrIdentifierExists
		}
		logger.Error().Err(err).Int64("userID", student.UserID).Str("identifier", student.Identifier).Msg("Error executing create student query")
		return fmt.Errorf("error creating student: %w", err)
	}

	logger.Info().Int64("userID", student.UserID).Str("identifier", student.Identifier).Msg("Student created successfully")
	return nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *StudentRepository) GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error) {
	var student models.Student
	sql, args, err := r.sb.Select("id", "user_id", "identifier", "department_id", "graduation_year").
		From("students").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get student by user ID SQL")
		return nil, fmt.Errorf("failed to build get student query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&student.ID, &student.UserID, &student.Identifier, &student.DepartmentID, &student.GraduationYear)

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

// IdentifierExists checks if a student identifier already exists
func (r *StudentRepository) IdentifierExists(ctx context.Context, identifier string) (bool, error) {
	var exists bool
	sql, args, err := r.sb.Select("1").
		From("students").
		Where(squirrel.Eq{"identifier": identifier}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building student identifier exists SQL")
		return false, fmt.Errorf("failed to build student identifier exists query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Str("identifier", identifier).Msg("Error checking student identifier existence")
		return false, fmt.Errorf("error checking student identifier existence: %w", err)
	}

	return exists, nil
}
