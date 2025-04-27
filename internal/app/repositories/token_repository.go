package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5" // Import pgx for ErrNoRows
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/dberrors" // Import the dberrors package
	"github.com/yigit/unisphere/internal/pkg/logger"   // Import logger
)

// TokenRepository handles token database operations
type TokenRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType // Add squirrel statement builder
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar), // Initialize squirrel
	}
}

// CreateToken creates a new refresh token
func (r *TokenRepository) CreateToken(ctx context.Context, token string, userID int64, expiryDate time.Time) error {
	sql, args, err := r.sb.Insert("refresh_tokens").
		Columns("token", "user_id", "expiry_date", "is_revoked", "created_at").
		Values(token, userID, expiryDate, false, time.Now()).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create token SQL")
		return fmt.Errorf("failed to build create token query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		// Use the generic function, assuming constraint name is "refresh_tokens_token_key"
		if dberrors.IsDuplicateConstraintError(err, "refresh_tokens_token_key") { // Use dberrors prefix
			logger.Warn().Str("token", token).Msg("Attempted to create duplicate token")
			// This shouldn't happen with unique tokens, but handle defensively
			return apperrors.ErrTokenInvalid
		}
		logger.Error().Err(err).Str("token", token).Int64("userID", userID).Msg("Error executing create token query")
		return fmt.Errorf("error creating token: %w", err)
	}

	return nil
}

// GetTokenByValue retrieves token information by value
func (r *TokenRepository) GetTokenByValue(ctx context.Context, token string) (int64, time.Time, bool, error) {
	var userID int64
	var expiryDate time.Time
	var isRevoked bool

	sql, args, err := r.sb.Select("user_id", "expiry_date", "is_revoked").
		From("refresh_tokens").
		Where(squirrel.Eq{"token": token}).
		Limit(1).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get token by value SQL")
		return 0, time.Time{}, false, fmt.Errorf("failed to build get token query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&userID, &expiryDate, &isRevoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, false, apperrors.ErrTokenNotFound
		}
		logger.Error().Err(err).Str("token", token).Msg("Error scanning token row")
		return 0, time.Time{}, false, fmt.Errorf("error retrieving token: %w", err)
	}

	// Check if token is revoked
	if isRevoked {
		return 0, time.Time{}, false, apperrors.ErrTokenRevoked
	}

	// Check token expiration
	if expiryDate.Before(time.Now()) {
		return 0, time.Time{}, false, apperrors.ErrTokenExpired
	}

	return userID, expiryDate, isRevoked, nil
}

// RevokeToken revokes a token
func (r *TokenRepository) RevokeToken(ctx context.Context, token string) error {
	sql, args, err := r.sb.Update("refresh_tokens").
		Set("is_revoked", true).
		Where(squirrel.Eq{"token": token}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building revoke token SQL")
		return fmt.Errorf("failed to build revoke token query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Str("token", token).Msg("Error executing revoke token query")
		return fmt.Errorf("error revoking token: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.ErrTokenNotFound
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	sql, args, err := r.sb.Update("refresh_tokens").
		Set("is_revoked", true).
		Where(squirrel.Eq{"user_id": userID, "is_revoked": false}). // Only revoke active tokens
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("userID", userID).Msg("Error building revoke all user tokens SQL")
		return fmt.Errorf("failed to build revoke all user tokens query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		// Don't return ErrNotFound here, as it's okay if the user had no active tokens.
		logger.Error().Err(err).Int64("userID", userID).Msg("Error executing revoke all user tokens query")
		return fmt.Errorf("error revoking user tokens: %w", err)
	}

	// Log how many were revoked? Maybe not necessary unless for debugging.
	// logger.Info().Int64("userID", userID).Int64("revokedCount", cmdTag.RowsAffected()).Msg("Revoked user tokens")

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	now := time.Now()

	sql, args, err := r.sb.Delete("refresh_tokens").
		Where(squirrel.Or{
			squirrel.Lt{"expiry_date": now}, // Expired tokens
			squirrel.And{ // Revoked tokens older than 30 days
				squirrel.Eq{"is_revoked": true},
				squirrel.Lt{"created_at": thirtyDaysAgo},
			},
		}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building cleanup tokens SQL")
		return 0, fmt.Errorf("failed to build cleanup tokens query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing cleanup tokens query")
		return 0, fmt.Errorf("error cleaning up tokens: %w", err)
	}

	deletedCount := cmdTag.RowsAffected()
	logger.Info().Int64("deletedCount", deletedCount).Msg("Cleaned up expired/old revoked tokens")

	return deletedCount, nil
}
