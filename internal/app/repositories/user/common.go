package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/dberrors"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// CommonRepository handles common user database operations
type CommonRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewRepository creates a new CommonRepository
func NewRepository(db *pgxpool.Pool) *CommonRepository {
	return &CommonRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// CreateUser creates a new user
func (r *CommonRepository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	columns := []string{"email", "password", "first_name", "last_name", "role_type", "is_active"}
	values := []interface{}{user.Email, user.Password, user.FirstName, user.LastName, user.RoleType, user.IsActive}

	// Add department_id if it exists
	if user.DepartmentID != nil {
		columns = append(columns, "department_id")
		values = append(values, *user.DepartmentID)
		logger.Info().Int64("departmentID", *user.DepartmentID).Msg("Adding department_id to user creation")
	}

	sqlBuilder := r.sb.Insert("users").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING id")

	sql, args, err := sqlBuilder.ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create user SQL")
		return 0, fmt.Errorf("failed to build create user query: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		if dberrors.IsDuplicateConstraintError(err, "users_email_key") {
			logger.Warn().Str("email", user.Email).Msg("Attempted to create user with duplicate email")
			return 0, apperrors.ErrEmailAlreadyExists
		}
		logger.Error().Err(err).Str("email", user.Email).Msg("Error executing create user query")
		return 0, fmt.Errorf("error creating user: %w", err)
	}

	// Log successful user creation with department information
	logEvent := logger.Info().Int64("userID", id).Str("email", user.Email)
	if user.DepartmentID != nil {
		logEvent = logEvent.Int64("departmentID", *user.DepartmentID)
	}
	logEvent.Msg("User created successfully")

	return id, nil
}

// GetUserByEmail retrieves a user by email
func (r *CommonRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	sql, args, err := r.sb.Select(
		"id", "email", "password", "first_name", "last_name",
		"created_at", "updated_at", "role_type", "is_active", "last_login_at",
		"department_id", "profile_photo_file_id",
	).
		From("users").
		Where(squirrel.Eq{"email": email}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get user by email SQL")
		return nil, fmt.Errorf("failed to build get user query: %w", err)
	}

	var lastLoginAt pgtype.Timestamp
	var departmentID pgtype.Int8
	var profilePhotoFileID pgtype.Int8
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive, &lastLoginAt,
		&departmentID, &profilePhotoFileID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Str("email", email).Msg("User not found by email")
			return nil, apperrors.ErrUserNotFound
		}
		logger.Error().Err(err).Str("email", email).Msg("Error scanning user row")
		return nil, fmt.Errorf("error retrieving user by email: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if departmentID.Valid {
		id := departmentID.Int64
		user.DepartmentID = &id
	}
	if profilePhotoFileID.Valid {
		id := profilePhotoFileID.Int64
		user.ProfilePhotoFileID = &id
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *CommonRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	sql, args, err := r.sb.Select(
		"id", "email", "password", "first_name", "last_name",
		"created_at", "updated_at", "role_type", "is_active", "last_login_at",
		"department_id", "profile_photo_file_id",
	).
		From("users").
		Where(squirrel.Eq{"id": id}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get user by ID SQL")
		return nil, fmt.Errorf("failed to build get user query: %w", err)
	}

	var lastLoginAt pgtype.Timestamp
	var departmentID pgtype.Int8
	var profilePhotoFileID pgtype.Int8
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive, &lastLoginAt,
		&departmentID, &profilePhotoFileID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Int64("userID", id).Msg("User not found by ID")
			return nil, apperrors.ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", id).Msg("Error scanning user row")
		return nil, fmt.Errorf("error retrieving user by ID: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if departmentID.Valid {
		id := departmentID.Int64
		user.DepartmentID = &id
	}
	if profilePhotoFileID.Valid {
		id := profilePhotoFileID.Int64
		user.ProfilePhotoFileID = &id
	}

	return &user, nil
}

// EmailExists checks if an email already exists
func (r *CommonRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	sql, args, err := r.sb.Select("1").
		From("users").
		Where(squirrel.Eq{"email": email}).
		Prefix("SELECT EXISTS (").
		Suffix(")").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building email exists SQL")
		return false, fmt.Errorf("failed to build email exists query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Str("email", email).Msg("Error checking email existence")
		return false, fmt.Errorf("error checking email existence: %w", err)
	}

	return exists, nil
}

// GetDepartmentNameByID retrieves department name by ID
func (r *CommonRepository) GetDepartmentNameByID(ctx context.Context, departmentID int64) (string, error) {
	var name string
	sql, args, err := r.sb.Select("name").
		From("departments").
		Where(squirrel.Eq{"id": departmentID}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get department name SQL")
		return "", fmt.Errorf("failed to build get department name query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Int64("departmentID", departmentID).Msg("Department not found by ID")
			return "", apperrors.ErrDepartmentNotFound
		}
		logger.Error().Err(err).Int64("departmentID", departmentID).Msg("Error scanning department name")
		return "", fmt.Errorf("error retrieving department name: %w", err)
	}

	return name, nil
}

// UpdateLastLogin updates the last login time
func (r *CommonRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	sql, args, err := r.sb.Update("users").
		Set("last_login_at", time.Now()).
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error building update last login SQL")
		return fmt.Errorf("failed to build update last login query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error executing update last login query")
		return fmt.Errorf("failed to update last login time: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn().Int64("userID", userID).Msg("Attempted to update last login for non-existent user")
		return apperrors.ErrUserNotFound
	}

	logger.Info().Int64("userID", userID).Msg("Last login time updated")
	return nil
}

// UpdateUserProfilePhotoURL updates only the profile photo URL for a given user.
// If photoURL is nil or empty, it sets the database column to NULL.
func (r *CommonRepository) UpdateUserProfilePhotoURL(ctx context.Context, userID int64, photoURL *string) error {
	var photoURLArg interface{}
	if photoURL != nil && *photoURL != "" {
		photoURLArg = *photoURL
	} else {
		photoURLArg = nil // Set to SQL NULL
	}

	sql, args, err := r.sb.Update("users").
		Set("profile_photo_url", photoURLArg).
		Set("updated_at", time.Now()). // Also update the updated_at timestamp
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error building update profile photo URL SQL")
		return fmt.Errorf("failed to build update profile photo URL query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error executing update profile photo URL query")
		return fmt.Errorf("failed to update profile photo URL: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn().Int64("userID", userID).Msg("Attempted to update profile photo URL for non-existent user")
		return apperrors.ErrUserNotFound
	}

	logger.Info().Int64("userID", userID).Msg("Profile photo URL updated")
	return nil
}

// UpdateUserProfilePhotoFileID updates the profile photo file ID for a given user.
func (r *CommonRepository) UpdateUserProfilePhotoFileID(ctx context.Context, userID int64, fileID *int64) error {
	var fileIDArg interface{}
	if fileID != nil {
		fileIDArg = *fileID
	} else {
		fileIDArg = nil // Set to SQL NULL
	}

	sql, args, err := r.sb.Update("users").
		Set("profile_photo_file_id", fileIDArg).
		Set("updated_at", time.Now()). // Also update the updated_at timestamp
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error building update profile photo file ID SQL")
		return fmt.Errorf("failed to build update profile photo file ID query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error executing update profile photo file ID query")
		return fmt.Errorf("failed to update profile photo file ID: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn().Int64("userID", userID).Msg("Attempted to update profile photo file ID for non-existent user")
		return apperrors.ErrUserNotFound
	}

	logger.Info().Int64("userID", userID).Msg("Profile photo file ID updated")
	return nil
}

// UpdateUserProfile updates a user's basic profile information (first name, last name, email)
func (r *CommonRepository) UpdateUserProfile(ctx context.Context, userID int64, firstName, lastName, email string) error {
	query := `
		UPDATE users 
		SET first_name = $2, last_name = $3, email = $4
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, userID, firstName, lastName, email)
	if err != nil {
		if dberrors.IsDuplicateConstraintError(err, "users_email_key") {
			logger.Warn().Str("email", email).Int64("userID", userID).Msg("Attempted to update user with duplicate email")
			return apperrors.ErrEmailAlreadyExists
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error updating user profile")
		return fmt.Errorf("error updating user profile: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		logger.Warn().Int64("userID", userID).Msg("User not found when updating profile")
		return apperrors.ErrUserNotFound
	}

	logger.Info().Int64("userID", userID).Msg("User profile updated successfully")
	return nil
}
