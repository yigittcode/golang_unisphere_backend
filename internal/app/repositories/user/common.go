package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// Common errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already in use")
)

// Repository handles common user database operations
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new Repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	// Check email availability
	exists, err := r.EmailExists(ctx, user.Email)
	if err != nil {
		return 0, fmt.Errorf("error checking email: %w", err)
	}
	if exists {
		return 0, ErrEmailAlreadyExists
	}

	var id int64
	err = r.db.QueryRow(ctx, `
		INSERT INTO users (email, password, first_name, last_name, role_type, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		user.Email, user.Password, user.FirstName, user.LastName, user.RoleType, user.IsActive).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating user: %w", err)
	}

	return id, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, role_type, is_active
		FROM users
		WHERE email = $1`,
		email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, role_type, is_active
		FROM users
		WHERE id = $1`,
		id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// EmailExists checks if an email already exists
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking email: %w", err)
	}

	return exists, nil
}

// GetDepartmentNameByID retrieves department name by ID
func (r *Repository) GetDepartmentNameByID(ctx context.Context, departmentID int64) (string, error) {
	var name string
	err := r.db.QueryRow(ctx, `
		SELECT name FROM departments WHERE id = $1`,
		departmentID).Scan(&name)

	if err != nil {
		return "", fmt.Errorf("department not found: %w", err)
	}

	return name, nil
}

// UpdateLastLogin updates the last login time
func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET last_login_at = $1
		WHERE id = $2`,
		now, userID)

	if err != nil {
		return fmt.Errorf("failed to update last login time: %w", err)
	}

	return nil
}
 