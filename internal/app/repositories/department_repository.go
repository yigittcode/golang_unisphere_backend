package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// Department error types
var (
	// ErrDepartmentNotFound is returned when a department is not found.
	ErrDepartmentNotFound = ErrNotFound // Use shared ErrNotFound
	// ErrDepartmentAlreadyExists is returned when a department with the same name or code exists.
	ErrDepartmentAlreadyExists = errors.New("department with this name or code already exists")
	// ErrDepartmentHasRelations is returned when trying to delete a department with associated data.
	ErrDepartmentHasRelations = errors.New("department has associated data and cannot be deleted")
)

// DepartmentRepository handles database operations for departments
type DepartmentRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create creates a new department
func (r *DepartmentRepository) Create(ctx context.Context, department *models.Department) error {
	sql, args, err := r.sb.Insert("departments").
		Columns("faculty_id", "name", "code").
		Values(department.FacultyID, department.Name, department.Code).
		Suffix("RETURNING id"). // Ensure ID is scanned back if needed, though current signature doesn't return it.
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create department SQL")
		return fmt.Errorf("failed to build create department query: %w", err)
	}

	// Scan the returned ID back into the department struct
	err = r.db.QueryRow(ctx, sql, args...).Scan(&department.ID)
	if err != nil {
		if isDuplicateKeyError(err) { // Assuming isDuplicateKeyError is defined or imported
			return ErrDepartmentAlreadyExists
		}
		logger.Error().Err(err).Msg("Error executing create department query")
		return fmt.Errorf("error creating department: %w", err)
	}

	return nil
}

// GetByID retrieves a department by ID
func (r *DepartmentRepository) GetByID(ctx context.Context, id int64) (*models.Department, error) {
	sql, args, err := r.sb.Select("id", "faculty_id", "name", "code").
		From("departments").
		Where(squirrel.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get department by ID SQL")
		return nil, fmt.Errorf("failed to build get department query: %w", err)
	}

	var department models.Department
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&department.ID,
		&department.FacultyID,
		&department.Name,
		&department.Code,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDepartmentNotFound
		}
		logger.Error().Err(err).Int64("departmentID", id).Msg("Error scanning department row")
		return nil, fmt.Errorf("error retrieving department: %w", err)
	}

	return &department, nil
}

// GetAll retrieves all departments
func (r *DepartmentRepository) GetAll(ctx context.Context) ([]*models.Department, error) {
	sql, args, err := r.sb.Select("id", "faculty_id", "name", "code").
		From("departments").
		OrderBy("name ASC").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get all departments SQL")
		return nil, fmt.Errorf("failed to build get all departments query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing get all departments query")
		return nil, fmt.Errorf("error querying departments: %w", err)
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
			logger.Error().Err(err).Msg("Error scanning department row during get all")
			return nil, fmt.Errorf("error scanning department row: %w", err)
		}
		departments = append(departments, &department)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating department rows")
		return nil, fmt.Errorf("error iterating department rows: %w", err)
	}

	return departments, nil
}

// GetByFacultyID retrieves all departments for a given faculty
func (r *DepartmentRepository) GetByFacultyID(ctx context.Context, facultyID int64) ([]*models.Department, error) {
	sql, args, err := r.sb.Select("id", "faculty_id", "name", "code").
		From("departments").
		Where(squirrel.Eq{"faculty_id": facultyID}).
		OrderBy("name ASC").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get departments by faculty ID SQL")
		return nil, fmt.Errorf("failed to build get departments by faculty query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("facultyID", facultyID).Msg("Error executing get departments by faculty query")
		return nil, fmt.Errorf("error querying departments by faculty: %w", err)
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
			logger.Error().Err(err).Msg("Error scanning department row during get by faculty")
			return nil, fmt.Errorf("error scanning department row: %w", err)
		}
		departments = append(departments, &department)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating department rows for faculty")
		return nil, fmt.Errorf("error iterating department rows by faculty: %w", err)
	}

	return departments, nil
}

// Update updates an existing department
func (r *DepartmentRepository) Update(ctx context.Context, department *models.Department) error {
	sql, args, err := r.sb.Update("departments").
		SetMap(map[string]interface{}{
			"faculty_id": department.FacultyID,
			"name":       department.Name,
			"code":       department.Code,
			// Assuming updated_at trigger exists
		}).
		Where(squirrel.Eq{"id": department.ID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building update department SQL")
		return fmt.Errorf("failed to build update department query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDepartmentAlreadyExists
		}
		logger.Error().Err(err).Int64("departmentID", department.ID).Msg("Error executing update department query")
		return fmt.Errorf("error updating department: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}

	return nil
}

// Delete deletes a department by ID
func (r *DepartmentRepository) Delete(ctx context.Context, id int64) error {
	// Check for related entities (example: past_exams, class_notes, students, instructors, courses)
	// A more robust check might involve querying multiple tables or using foreign key constraints.
	relatedTables := []string{"past_exams", "class_notes", "students", "instructors", "courses"}
	for _, table := range relatedTables {
		var exists bool
		checkSql, checkArgs, err := r.sb.Select("1").
			From(table).
			Where(squirrel.Eq{"department_id": id}).
			Prefix("SELECT EXISTS (").Suffix(")").
			Limit(1).
			ToSql()

		if err != nil {
			logger.Error().Err(err).Str("table", table).Msg("Error building check related entities SQL")
			return fmt.Errorf("failed to build check for related %s: %w", table, err)
		}

		err = r.db.QueryRow(ctx, checkSql, checkArgs...).Scan(&exists)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			logger.Error().Err(err).Int64("departmentID", id).Str("table", table).Msg("Error checking related entities")
			return fmt.Errorf("error checking related %s: %w", table, err)
		}
		if exists {
			logger.Warn().Int64("departmentID", id).Str("table", table).Msg("Attempted to delete department with related data")
			return ErrDepartmentHasRelations
		}
	}

	// Proceed with deletion
	sql, args, err := r.sb.Delete("departments").
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building delete department SQL")
		return fmt.Errorf("failed to build delete department query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("departmentID", id).Msg("Error executing delete department query")
		return fmt.Errorf("error deleting department: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDepartmentNotFound
	}

	return nil
}

// DepartmentExistsByNameOrCode likely redundant due to unique constraints & create/update error handling
// Kept for reference or potential other uses.
func (r *DepartmentRepository) DepartmentExistsByNameOrCode(ctx context.Context, name, code string) (bool, error) {
	sql, args, err := r.sb.Select("1").
		From("departments").
		Where(squirrel.Or{squirrel.Eq{"name": name}, squirrel.Eq{"code": code}}).
		Prefix("SELECT EXISTS (").Suffix(")").
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building department exists SQL")
		return false, fmt.Errorf("failed to build department existence query: %w", err)
	}

	var exists bool
	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		logger.Error().Err(err).Str("name", name).Str("code", code).Msg("Error checking department existence")
		return false, fmt.Errorf("error checking department existence: %w", err)
	}

	return exists, nil
}

// isDuplicateKeyError needs to be accessible (e.g., defined here or in a shared util)
// Reusing the one potentially defined in faculty_repository.go for now, assuming it's moved/
// or duplicating it here temporarily.
// func isDuplicateKeyError(err error) bool {
// 	var pgErr *pgconn.PgError
// 	return errors.As(err, &pgErr) && pgErr.Code == "23505"
// }
