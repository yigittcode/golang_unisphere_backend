package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// PasswordResetTokenRepository manages password reset tokens in the database
type PasswordResetTokenRepository struct {
	db *pgxpool.Pool
}

// NewPasswordResetTokenRepository creates a new PasswordResetTokenRepository
func NewPasswordResetTokenRepository(db *pgxpool.Pool) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{
		db: db,
	}
}

// CreateToken stores a new password reset token in the database
func (r *PasswordResetTokenRepository) CreateToken(ctx context.Context, userID int64, token string, expiryDate time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expiry_date)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, token, expiryDate)
	if err != nil {
		return fmt.Errorf("error creating password reset token: %w", err)
	}

	return nil
}

// GetTokenInfo retrieves user ID and expiry date for a given token
func (r *PasswordResetTokenRepository) GetTokenInfo(ctx context.Context, token string) (int64, time.Time, bool, error) {
	query := `
		SELECT user_id, expiry_date, used
		FROM password_reset_tokens
		WHERE token = $1
	`

	var userID int64
	var expiryDate time.Time
	var used bool

	err := r.db.QueryRow(ctx, query, token).Scan(&userID, &expiryDate, &used)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, time.Time{}, false, apperrors.ErrTokenNotFound
		}
		return 0, time.Time{}, false, fmt.Errorf("error retrieving password reset token: %w", err)
	}

	return userID, expiryDate, used, nil
}

// MarkTokenAsUsed marks a token as used to prevent reuse
func (r *PasswordResetTokenRepository) MarkTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE token = $1
	`

	result, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("error marking token as used: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.ErrTokenNotFound
	}

	return nil
}

// DeleteToken removes a token from the database
func (r *PasswordResetTokenRepository) DeleteToken(ctx context.Context, token string) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE token = $1
	`

	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("error deleting password reset token: %w", err)
	}

	return nil
}

// DeleteTokensByUserID removes all tokens for a specific user
func (r *PasswordResetTokenRepository) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error deleting password reset tokens for user: %w", err)
	}

	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (r *PasswordResetTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expiry_date < $1
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("error deleting expired password reset tokens: %w", err)
	}

	return nil
}
