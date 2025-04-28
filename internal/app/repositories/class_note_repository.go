package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// ClassNoteDetails includes detailed information about a class note, joining related tables.
type ClassNoteDetails struct {
	ID                int64           `db:"id" json:"id"`
	Year              int             `db:"year" json:"year"`
	Term              models.Term     `db:"term" json:"term"`
	DepartmentID      int64           `db:"department_id" json:"departmentId"`
	DepartmentName    string          `db:"department_name" json:"departmentName"`
	DepartmentCode    string          `db:"department_code" json:"departmentCode"`
	FacultyID         int64           `db:"faculty_id" json:"facultyId"`
	FacultyName       string          `db:"faculty_name" json:"facultyName"`
	FacultyCode       string          `db:"faculty_code" json:"facultyCode"`
	CourseCode        string          `db:"course_code" json:"courseCode"`
	Title             string          `db:"title" json:"title"`
	Content           string          `db:"content" json:"content"`
	UserID            int64           `db:"user_id" json:"userId"`
	UploaderFirstName string          `db:"uploader_first_name" json:"uploaderFirstName"`
	UploaderLastName  string          `db:"uploader_last_name" json:"uploaderLastName"`
	UploaderEmail     string          `db:"uploader_email" json:"uploaderEmail"`
	UploaderRole      models.RoleType `db:"uploader_role" json:"uploaderRole"`
	UploadedByStudent bool            `json:"uploadedByStudent"`
	CreatedAt         time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updatedAt"`
}

// GetAllNotesParams holds parameters for filtering and pagination.
type GetAllNotesParams struct {
	FacultyID    *int64
	DepartmentID *int64
	CourseCode   *string
	Year         *int
	Term         *models.Term
	SortBy       string
	SortOrder    string
	Page         int
	Size         int
}

// ClassNoteRepository handles database operations for ClassNote.
type ClassNoteRepository struct {
	DB *pgxpool.Pool
}

// NewClassNoteRepository creates a new instance of ClassNoteRepository.
func NewClassNoteRepository(db *pgxpool.Pool) *ClassNoteRepository {
	return &ClassNoteRepository{DB: db}
}

// Common select query builder for class notes with joins
func (r *ClassNoteRepository) selectClassNoteDetailsQuery() squirrel.SelectBuilder {
	return squirrel.Select(
		"cn.id", "cn.year", "cn.term", "cn.department_id", "d.name as department_name",
		"d.code as department_code", "d.faculty_id", "f.name as faculty_name", "f.code as faculty_code",
		"cn.course_code", "cn.title", "cn.content", "cn.user_id",
		"u.first_name as uploader_first_name", "u.last_name as uploader_last_name", "u.email as uploader_email", "u.role_type as uploader_role",
		"cn.created_at", "cn.updated_at",
	).From("class_notes cn").
		Join("departments d ON cn.department_id = d.id").
		Join("faculties f ON d.faculty_id = f.id").
		Join("users u ON cn.user_id = u.id").
		PlaceholderFormat(squirrel.Dollar)
}

// ScanClassNoteDetails scans a row into ClassNoteDetails struct.
func ScanClassNoteDetails(row pgx.Row) (*ClassNoteDetails, error) {
	var note ClassNoteDetails
	err := row.Scan(
		&note.ID, &note.Year, &note.Term, &note.DepartmentID, &note.DepartmentName,
		&note.DepartmentCode, &note.FacultyID, &note.FacultyName, &note.FacultyCode,
		&note.CourseCode, &note.Title, &note.Content, &note.UserID,
		&note.UploaderFirstName, &note.UploaderLastName, &note.UploaderEmail, &note.UploaderRole,
		&note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrClassNoteNotFound
		}
		logger.Error().Err(err).Msg("Error scanning class note details")
		return nil, err
	}
	note.UploadedByStudent = (note.UploaderRole == models.RoleStudent)
	return &note, nil
}

// CreateClassNote inserts a new class note into the database.
func (r *ClassNoteRepository) CreateClassNote(ctx context.Context, note *models.ClassNote) (int64, error) {
	sql, args, err := squirrel.Insert("class_notes").
		Columns("year", "term", "department_id", "course_code", "title", "content", "user_id").
		Values(note.Year, note.Term, note.DepartmentID, note.CourseCode, note.Title, note.Content, note.UserID).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create class note SQL")
		return 0, err
	}

	var id int64
	err = r.DB.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		// TODO: Check for specific constraint violations if needed
		logger.Error().Err(err).Msg("Error executing create class note query")
		return 0, err
	}

	return id, nil
}

// GetClassNoteByID retrieves a single class note by its ID with details.
func (r *ClassNoteRepository) GetClassNoteByID(ctx context.Context, id int64) (*ClassNoteDetails, error) {
	sqlBuilder := r.selectClassNoteDetailsQuery().Where(squirrel.Eq{"cn.id": id})
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building get class note by ID SQL")
		return nil, err
	}

	row := r.DB.QueryRow(ctx, sqlStr, args...)
	return ScanClassNoteDetails(row)
}

// GetAllClassNotes retrieves a paginated and filtered list of class notes with details.
func (r *ClassNoteRepository) GetAllClassNotes(ctx context.Context, params GetAllNotesParams) ([]*ClassNoteDetails, dto.PaginationInfo, error) {
	sqlBuilder := r.selectClassNoteDetailsQuery()
	countBuilder := squirrel.Select("count(*)").From("class_notes cn").
		Join("departments d ON cn.department_id = d.id"). // Need join for faculty filter
		PlaceholderFormat(squirrel.Dollar)

	// Apply filters
	if params.FacultyID != nil {
		// Join with faculties needed if filtering by faculty ID
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"d.faculty_id": *params.FacultyID})
		countBuilder = countBuilder.Where(squirrel.Eq{"d.faculty_id": *params.FacultyID})
	}
	if params.DepartmentID != nil {
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"cn.department_id": *params.DepartmentID})
		countBuilder = countBuilder.Where(squirrel.Eq{"cn.department_id": *params.DepartmentID})
	}
	if params.CourseCode != nil && *params.CourseCode != "" {
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"cn.course_code": *params.CourseCode})
		countBuilder = countBuilder.Where(squirrel.Eq{"cn.course_code": *params.CourseCode})
	}
	if params.Year != nil {
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"cn.year": *params.Year})
		countBuilder = countBuilder.Where(squirrel.Eq{"cn.year": *params.Year})
	}
	if params.Term != nil {
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"cn.term": *params.Term})
		countBuilder = countBuilder.Where(squirrel.Eq{"cn.term": *params.Term})
	}

	// Get total count
	countSql, countArgs, err := countBuilder.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building count query SQL")
		return nil, dto.PaginationInfo{}, err
	}

	var totalItems int64
	err = r.DB.QueryRow(ctx, countSql, countArgs...).Scan(&totalItems)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing count query")
		return nil, dto.PaginationInfo{}, err
	}

	// Use helper to create pagination info
	pagnation := helpers.NewPaginationInfo(totalItems, params.Page, params.Size)

	if totalItems == 0 {
		// Return the created pagination struct (value)
		return []*ClassNoteDetails{}, pagnation, nil
	}

	// Use helper to calculate offset and limit
	offset, limit := helpers.CalculateOffsetLimit(params.Page, params.Size)

	// Apply sorting
	sortBy := "cn.created_at" // Default sort
	if params.SortBy != "" {
		// Basic validation to prevent SQL injection
		allowedSorts := map[string]string{
			"createdAt": "cn.created_at",
			"year":      "cn.year",
			"term":      "cn.term",
			"course":    "cn.course_code",
			"title":     "cn.title",
		}
		if validSort, ok := allowedSorts[params.SortBy]; ok {
			sortBy = validSort
		}
	}
	sortOrder := "DESC" // Default order
	if strings.ToUpper(params.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}
	sqlBuilder = sqlBuilder.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination using calculated limit and offset
	sqlBuilder = sqlBuilder.Limit(uint64(limit)).Offset(offset)

	// Execute main query
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building get all class notes SQL")
		return nil, dto.PaginationInfo{}, err
	}

	rows, err := r.DB.Query(ctx, sqlStr, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing get all class notes query")
		return nil, dto.PaginationInfo{}, err
	}
	defer rows.Close()

	notes := make([]*ClassNoteDetails, 0)
	for rows.Next() {
		note, err := ScanClassNoteDetails(rows)
		if err != nil {
			// Log error but continue processing other rows
			logger.Error().Err(err).Msg("Error scanning one class note during get all")
			continue
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error after iterating through class note rows")
		// Return pagination info even if iteration fails partially?
		// Returning the calculated pagination value here might be reasonable.
		return nil, pagnation, fmt.Errorf("database iteration error: %w", err)
	}

	// Return the created pagination struct (value)
	return notes, pagnation, nil
}

// UpdateClassNote updates an existing class note.
func (r *ClassNoteRepository) UpdateClassNote(ctx context.Context, note *models.ClassNote) error {
	sql, args, err := squirrel.Update("class_notes").
		Set("year", note.Year).
		Set("term", note.Term).
		Set("department_id", note.DepartmentID).
		Set("course_code", note.CourseCode).
		Set("title", note.Title).
		Set("content", note.Content).
		// updated_at is handled by trigger
		Where(squirrel.Eq{"id": note.ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building update class note SQL")
		return err
	}

	cmdTag, err := r.DB.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing update class note query")
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.ErrClassNoteNotFound // Or a more specific error like ErrUpdateFailed
	}

	return nil
}

// DeleteClassNote deletes a class note by its ID.
func (r *ClassNoteRepository) DeleteClassNote(ctx context.Context, id int64) error {
	sql, args, err := squirrel.Delete("class_notes").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building delete class note SQL")
		return err
	}

	cmdTag, err := r.DB.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing delete class note query")
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.ErrClassNoteNotFound // Note not found or already deleted
	}

	return nil
}

// GetClassNotesByUploaderID retrieves all class notes uploaded by a specific user.
func (r *ClassNoteRepository) GetClassNotesByUploaderID(ctx context.Context, userID int64) ([]*ClassNoteDetails, error) {
	sqlBuilder := r.selectClassNoteDetailsQuery().Where(squirrel.Eq{"cn.user_id": userID}).OrderBy("cn.created_at DESC")
	sqlStr, args, err := sqlBuilder.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building get class notes by uploader ID SQL")
		return nil, err
	}

	rows, err := r.DB.Query(ctx, sqlStr, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing get class notes by uploader ID query")
		return nil, err
	}
	defer rows.Close()

	notes := make([]*ClassNoteDetails, 0)
	for rows.Next() {
		note, err := ScanClassNoteDetails(rows)
		if err != nil {
			logger.Error().Err(err).Msg("Error scanning one class note during get by uploader")
			continue // Log and skip
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating through class note rows for uploader")
		return nil, err
	}

	return notes, nil
}

// This error has been moved to apperrors package
// Use apperrors.ErrNotFound or more specific errors like apperrors.ErrClassNoteNotFound instead

// GetFileRepository returns the file repository
func (r *ClassNoteRepository) GetFileRepository() *FileRepository {
	return NewFileRepository(r.DB)
}
