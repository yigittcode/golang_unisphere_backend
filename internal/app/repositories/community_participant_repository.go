package repositories

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// CommunityParticipantRepository handles database operations for community participants
type CommunityParticipantRepository struct {
	db *pgxpool.Pool
}

// NewCommunityParticipantRepository creates a new CommunityParticipantRepository
func NewCommunityParticipantRepository(db *pgxpool.Pool) *CommunityParticipantRepository {
	return &CommunityParticipantRepository{db: db}
}

// GetParticipantsByCommunityID retrieves all participants for a specific community
func (r *CommunityParticipantRepository) GetParticipantsByCommunityID(ctx context.Context, communityID int64) ([]*models.CommunityParticipant, error) {
	query := squirrel.Select(
		"cp.id", "cp.community_id", "cp.user_id", "cp.joined_at",
	).
		From("community_participants cp").
		Where("cp.community_id = ?", communityID).
		OrderBy("cp.joined_at DESC").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var participants []*models.CommunityParticipant
	for rows.Next() {
		var participant models.CommunityParticipant
		err := rows.Scan(
			&participant.ID,
			&participant.CommunityID,
			&participant.UserID,
			&participant.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		participants = append(participants, &participant)
	}

	return participants, nil
}

// GetParticipantCountByCommunityID retrieves the number of participants for a specific community
func (r *CommunityParticipantRepository) GetParticipantCountByCommunityID(ctx context.Context, communityID int64) (int, error) {
	query := squirrel.Select("COUNT(*)").
		From("community_participants").
		Where("community_id = ?", communityID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("error building SQL: %w", err)
	}

	var count int
	err = r.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	return count, nil
}

// GetParticipantCountsByCommunityIDs retrieves the number of participants for multiple communities
func (r *CommunityParticipantRepository) GetParticipantCountsByCommunityIDs(ctx context.Context, communityIDs []int64) (map[int64]int, error) {
	if len(communityIDs) == 0 {
		return make(map[int64]int), nil
	}

	query := squirrel.Select("community_id", "COUNT(*)").
		From("community_participants").
		Where(squirrel.Eq{"community_id": communityIDs}).
		GroupBy("community_id").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var communityID int64
		var count int
		if err := rows.Scan(&communityID, &count); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		counts[communityID] = count
	}

	return counts, nil
}

// GetCommunitiesByUserID retrieves all communities a user is participating in
func (r *CommunityParticipantRepository) GetCommunitiesByUserID(ctx context.Context, userID int64) ([]int64, error) {
	query := squirrel.Select("community_id").
		From("community_participants").
		Where("user_id = ?", userID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var communityIDs []int64
	for rows.Next() {
		var communityID int64
		err := rows.Scan(&communityID)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		communityIDs = append(communityIDs, communityID)
	}

	return communityIDs, nil
}

// IsUserParticipant checks if a user is a participant in a specific community
func (r *CommunityParticipantRepository) IsUserParticipant(ctx context.Context, communityID, userID int64) (bool, error) {
	query := squirrel.Select("1").
		From("community_participants").
		Where("community_id = ? AND user_id = ?", communityID, userID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("error building SQL: %w", err)
	}

	var exists int
	err = r.db.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error executing query: %w", err)
	}

	return true, nil
}

// AddParticipant adds a user as a participant to a community
func (r *CommunityParticipantRepository) AddParticipant(ctx context.Context, communityID, userID int64) (int64, error) {
	query := squirrel.Insert("community_participants").
		Columns("community_id", "user_id").
		Values(communityID, userID).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("error building SQL: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	return id, nil
}

// RemoveParticipant removes a user as a participant from a community
func (r *CommunityParticipantRepository) RemoveParticipant(ctx context.Context, communityID, userID int64) error {
	query := squirrel.Delete("community_participants").
		Where("community_id = ? AND user_id = ?", communityID, userID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
