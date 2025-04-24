package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Token errors
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenRevoked  = errors.New("token revoked")
)

// TokenRepository handles token database operations
type TokenRepository struct {
	db *pgxpool.Pool
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// CreateToken creates a new refresh token
func (r *TokenRepository) CreateToken(ctx context.Context, token string, userID int64, expiryDate time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (token, user_id, expiry_date, is_revoked, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		token, userID, expiryDate, false, time.Now())

	if err != nil {
		return fmt.Errorf("error creating token: %w", err)
	}

	return nil
}

// GetTokenByValue retrieves token information by value
func (r *TokenRepository) GetTokenByValue(ctx context.Context, token string) (int64, time.Time, bool, error) {
	var userID int64
	var expiryDate time.Time
	var isRevoked bool

	err := r.db.QueryRow(ctx, `
		SELECT user_id, expiry_date, is_revoked
		FROM refresh_tokens
		WHERE token = $1`,
		token).Scan(&userID, &expiryDate, &isRevoked)

	if err != nil {
		return 0, time.Time{}, false, ErrTokenNotFound
	}

	// Check if token is revoked
	if isRevoked {
		return 0, time.Time{}, false, ErrTokenRevoked
	}

	// Check token expiration
	if expiryDate.Before(time.Now()) {
		return 0, time.Time{}, false, ErrTokenExpired
	}

	return userID, expiryDate, isRevoked, nil
}

// RevokeToken revokes a token
func (r *TokenRepository) RevokeToken(ctx context.Context, token string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = true
		WHERE token = $1`,
		token)

	if err != nil {
		return fmt.Errorf("error revoking token: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = true
		WHERE user_id = $1 AND is_revoked = false`,
		userID)

	if err != nil {
		return fmt.Errorf("error revoking user tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE expiry_date < $1 OR (is_revoked = true AND created_at < $2)`,
		time.Now(), time.Now().Add(-30*24*time.Hour)) // Remove revoked tokens older than 30 days

	if err != nil {
		return 0, fmt.Errorf("error cleaning up tokens: %w", err)
	}

	return result.RowsAffected(), nil
}
