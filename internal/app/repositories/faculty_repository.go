package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// Faculty error types
var (
	// ErrFacultyNotFound is returned when a faculty is not found.
	ErrFacultyNotFound = ErrNotFound // Use shared ErrNotFound
	// ErrFacultyAlreadyExists is returned when a faculty with the same name or code exists.
	ErrFacultyAlreadyExists = errors.New("faculty with this name or code already exists")
	// ErrFacultyHasDepartments is returned when trying to delete a faculty with associated departments.
	ErrFacultyHasDepartments = errors.New("faculty has associated departments and cannot be deleted")
)

// FacultyRepository handles faculty database operations
type FacultyRepository struct {
	db *pgxpool.Pool
	// Use squirrel instance with placeholder format
	sb squirrel.StatementBuilderType
}

// NewFacultyRepository creates a new FacultyRepository
func NewFacultyRepository(db *pgxpool.Pool) *FacultyRepository {
	return &FacultyRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// isDuplicateKeyError checks if the error is a PostgreSQL unique violation error.
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" // 23505 is unique_violation
}

// CreateFaculty creates a new faculty
func (r *FacultyRepository) CreateFaculty(ctx context.Context, faculty *models.Faculty) (int64, error) {
	sql, args, err := r.sb.Insert("faculties").
		Columns("name", "code", "description").
		Values(faculty.Name, faculty.Code, faculty.Description).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create faculty SQL")
		return 0, fmt.Errorf("failed to build create faculty query: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, ErrFacultyAlreadyExists
		}
		logger.Error().Err(err).Msg("Error executing create faculty query")
		return 0, fmt.Errorf("error creating faculty: %w", err)
	}

	return id, nil
}

// GetFacultyByID retrieves a faculty by ID
func (r *FacultyRepository) GetFacultyByID(ctx context.Context, id int64) (*models.Faculty, error) {
	sql, args, err := r.sb.Select("id", "name", "code", "description").
		From("faculties").
		Where(squirrel.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get faculty by ID SQL")
		return nil, fmt.Errorf("failed to build get faculty query: %w", err)
	}

	faculty := &models.Faculty{}
	err = r.db.QueryRow(ctx, sql, args...).Scan(&faculty.ID, &faculty.Name, &faculty.Code, &faculty.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFacultyNotFound // Use shared ErrNotFound
		}
		logger.Error().Err(err).Int64("facultyID", id).Msg("Error scanning faculty row")
		return nil, fmt.Errorf("error getting faculty by ID: %w", err)
	}

	return faculty, nil
}

// GetAllFaculties retrieves all faculties
func (r *FacultyRepository) GetAllFaculties(ctx context.Context) ([]*models.Faculty, error) {
	sql, args, err := r.sb.Select("id", "name", "code", "description").
		From("faculties").
		OrderBy("name ASC").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get all faculties SQL")
		return nil, fmt.Errorf("failed to build get all faculties query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing get all faculties query")
		return nil, fmt.Errorf("error querying faculties: %w", err)
	}
	defer rows.Close()

	faculties := []*models.Faculty{}
	for rows.Next() {
		faculty := &models.Faculty{}
		if err := rows.Scan(&faculty.ID, &faculty.Name, &faculty.Code, &faculty.Description); err != nil {
			logger.Error().Err(err).Msg("Error scanning faculty row during get all")
			// Decide whether to return partial list or error out
			return nil, fmt.Errorf("error scanning faculty row: %w", err)
		}
		faculties = append(faculties, faculty)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating faculty rows")
		return nil, fmt.Errorf("error iterating faculty rows: %w", err)
	}

	return faculties, nil
}

// UpdateFaculty updates an existing faculty
func (r *FacultyRepository) UpdateFaculty(ctx context.Context, faculty *models.Faculty) error {
	sql, args, err := r.sb.Update("faculties").
		SetMap(map[string]interface{}{
			"name":        faculty.Name,
			"code":        faculty.Code,
			"description": faculty.Description,
			// updated_at is not explicitly managed here, assuming a trigger or manual update elsewhere if needed
		}).
		Where(squirrel.Eq{"id": faculty.ID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building update faculty SQL")
		return fmt.Errorf("failed to build update faculty query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			// Attempted to update to a name/code that already exists
			return ErrFacultyAlreadyExists
		}
		logger.Error().Err(err).Int64("facultyID", faculty.ID).Msg("Error executing update faculty query")
		return fmt.Errorf("error updating faculty: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		// ID did not exist
		return ErrFacultyNotFound
	}

	return nil
}

// DeleteFaculty deletes a faculty by ID
func (r *FacultyRepository) DeleteFaculty(ctx context.Context, id int64) error {
	// Check if faculty has associated departments BEFORE deleting
	var hasDepartments bool
	checkSql, checkArgs, err := r.sb.Select("1").
		From("departments").
		Where(squirrel.Eq{"faculty_id": id}).
		Prefix("SELECT EXISTS (").Suffix(")").
		Limit(1). // Important for EXISTS performance
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building check departments SQL")
		return fmt.Errorf("failed to build check departments query: %w", err)
	}

	err = r.db.QueryRow(ctx, checkSql, checkArgs...).Scan(&hasDepartments)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) { // ErrNoRows is ok here, means false
		logger.Error().Err(err).Int64("facultyID", id).Msg("Error checking associated departments")
		return fmt.Errorf("error checking associated departments: %w", err)
	}

	if hasDepartments {
		return ErrFacultyHasDepartments
	}

	// Proceed with deletion
	sql, args, err := r.sb.Delete("faculties").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building delete faculty SQL")
		return fmt.Errorf("failed to build delete faculty query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("facultyID", id).Msg("Error executing delete faculty query")
		return fmt.Errorf("error deleting faculty: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		// Faculty not found (might have been deleted between check and delete)
		return ErrFacultyNotFound
	}

	return nil
}

// FacultyExistsByNameOrCode is likely redundant if handled by constraints/CreateFaculty error check
// but kept for potential use elsewhere or if constraints are not reliable.
func (r *FacultyRepository) FacultyExistsByNameOrCode(ctx context.Context, name, code string) (bool, error) {
	sql, args, err := r.sb.Select("1").
		From("faculties").
		Where(squirrel.Or{squirrel.Eq{"name": name}, squirrel.Eq{"code": code}}).
		Prefix("SELECT EXISTS (").Suffix(")").
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building faculty exists SQL")
		return false, fmt.Errorf("failed to build faculty existence query: %w", err)
	}

	var exists bool
	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) { // ErrNoRows is ok here, means false
		logger.Error().Err(err).Str("name", name).Str("code", code).Msg("Error checking faculty existence")
		return false, fmt.Errorf("error checking faculty existence: %w", err)
	}

	return exists, nil
}
