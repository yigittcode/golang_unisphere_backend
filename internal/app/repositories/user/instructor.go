package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/dberrors"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// InstructorRepository handles instructor database operations
type InstructorRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewInstructorRepository creates a new InstructorRepository
func NewInstructorRepository(db *pgxpool.Pool) *InstructorRepository {
	return &InstructorRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// CreateInstructor creates a new instructor
func (r *InstructorRepository) CreateInstructor(ctx context.Context, instructor *models.Instructor) error {
	sql, args, err := r.sb.Insert("instructors").
		Columns("user_id", "department_id", "title").
		Values(instructor.UserID, instructor.DepartmentID, instructor.Title).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create instructor SQL")
		return fmt.Errorf("failed to build create instructor query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		if dberrors.IsDuplicateConstraintError(err, "instructors_user_id_key") {
			logger.Warn().Int64("userID", instructor.UserID).Msg("Attempted to create duplicate instructor entry")
			return fmt.Errorf("instructor entry for this user already exists")
		}
		logger.Error().Err(err).Int64("userID", instructor.UserID).Msg("Error executing create instructor query")
		return fmt.Errorf("error creating instructor: %w", err)
	}

	logger.Info().Int64("userID", instructor.UserID).Msg("Instructor created successfully")
	return nil
}

// GetInstructorByUserID retrieves an instructor by user ID
func (r *InstructorRepository) GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error) {
	var instructor models.Instructor
	sql, args, err := r.sb.Select("id", "user_id", "department_id", "title").
		From("instructors").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get instructor by user ID SQL")
		return nil, fmt.Errorf("failed to build get instructor query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&instructor.ID, &instructor.UserID, &instructor.DepartmentID, &instructor.Title)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Int64("userID", userID).Msg("Instructor not found by user ID")
			return nil, apperrors.ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error scanning instructor row")
		return nil, fmt.Errorf("error retrieving instructor: %w", err)
	}

	return &instructor, nil
}

// GetInstructorsByDepartmentID retrieves all instructors in a department, including user details
func (r *InstructorRepository) GetInstructorsByDepartmentID(ctx context.Context, departmentID int64) ([]*models.Instructor, error) {
	sql, args, err := r.sb.Select(
		"i.id", "i.user_id", "i.department_id", "i.title",
		"u.first_name", "u.last_name", "u.email",
	).
		From("instructors i").
		Join("users u ON i.user_id = u.id").
		Where(squirrel.Eq{"i.department_id": departmentID}).
		OrderBy("u.last_name", "u.first_name").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("departmentID", departmentID).Msg("Error building get instructors by department SQL")
		return nil, fmt.Errorf("failed to build get instructors query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("departmentID", departmentID).Msg("Error executing get instructors by department query")
		return nil, fmt.Errorf("error querying instructors: %w", err)
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
			logger.Error().Err(err).Int64("departmentID", departmentID).Msg("Error scanning instructor row during department fetch")
			return nil, fmt.Errorf("error scanning instructor: %w", err)
		}
		instructors = append(instructors, instructor)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Int64("departmentID", departmentID).Msg("Error iterating instructor rows")
		return nil, fmt.Errorf("error iterating instructors: %w", err)
	}

	logger.Info().Int64("departmentID", departmentID).Int("count", len(instructors)).Msg("Successfully fetched instructors by department")
	return instructors, nil
}

// UpdateInstructorTitle updates an instructor's title
func (r *InstructorRepository) UpdateInstructorTitle(ctx context.Context, userID int64, newTitle string) error {
	sql, args, err := r.sb.Update("instructors").
		Set("title", newTitle).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error building update instructor title SQL")
		return fmt.Errorf("failed to build update instructor title query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Str("newTitle", newTitle).Msg("Error executing update instructor title query")
		return fmt.Errorf("error updating instructor title: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn().Int64("userID", userID).Msg("Attempted to update title for non-existent instructor/user")
		return apperrors.ErrUserNotFound
	}

	logger.Info().Int64("userID", userID).Msg("Instructor title updated successfully")
	return nil
}
