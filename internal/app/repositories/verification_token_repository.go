package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VerificationTokenRepository handles database operations for email verification tokens
type VerificationTokenRepository struct {
	db *pgxpool.Pool
}

// NewVerificationTokenRepository creates a new VerificationTokenRepository
func NewVerificationTokenRepository(db *pgxpool.Pool) *VerificationTokenRepository {
	return &VerificationTokenRepository{db: db}
}

// CreateToken creates a new email verification token for a user
func (r *VerificationTokenRepository) CreateToken(ctx context.Context, userID int64, token string, expiryDate time.Time) error {
	query := squirrel.Insert("email_verification_tokens").
		Columns("user_id", "token", "expiry_date").
		Values(userID, token, expiryDate).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error creating verification token: %w", err)
	}

	return nil
}

// GetTokenInfo retrieves token information by token value
func (r *VerificationTokenRepository) GetTokenInfo(ctx context.Context, token string) (int64, time.Time, error) {
	query := squirrel.Select("user_id", "expiry_date").
		From("email_verification_tokens").
		Where("token = ?", token).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("error building SQL: %w", err)
	}

	var userID int64
	var expiryDate time.Time

	err = r.db.QueryRow(ctx, sql, args...).Scan(&userID, &expiryDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, time.Time{}, fmt.Errorf("token not found")
		}
		return 0, time.Time{}, fmt.Errorf("error getting token info: %w", err)
	}

	return userID, expiryDate, nil
}

// DeleteToken deletes a verification token
func (r *VerificationTokenRepository) DeleteToken(ctx context.Context, token string) error {
	query := squirrel.Delete("email_verification_tokens").
		Where("token = ?", token).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}

	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (r *VerificationTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := squirrel.Delete("email_verification_tokens").
		Where("expiry_date < ?", time.Now()).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error deleting expired tokens: %w", err)
	}

	return nil
}

// DeleteTokensByUserID deletes all tokens for a specific user
func (r *VerificationTokenRepository) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	query := squirrel.Delete("email_verification_tokens").
		Where("user_id = ?", userID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error deleting tokens for user: %w", err)
	}

	return nil
}