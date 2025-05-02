package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// CommunityRepository handles database operations for communities
type CommunityRepository struct {
	db *pgxpool.Pool
}

// NewCommunityRepository creates a new CommunityRepository
func NewCommunityRepository(db *pgxpool.Pool) *CommunityRepository {
	return &CommunityRepository{db: db}
}

// GetAll retrieves all communities with filtering and pagination
func (r *CommunityRepository) GetAll(ctx context.Context, leadID *int64, search *string, page, pageSize int) ([]models.Community, int64, error) {
	// Build base query with table aliases
	query := squirrel.Select(
		"c.id", "c.name", "c.abbreviation", "c.lead_id", "c.profile_photo_file_id", "c.created_at", "c.updated_at",
	).
		From("communities c").
		PlaceholderFormat(squirrel.Dollar)

	// Add filters
	if leadID != nil {
		query = query.Where("c.lead_id = ?", *leadID)
	}
	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query = query.Where("(c.name ILIKE ? OR c.abbreviation ILIKE ?)", searchPattern, searchPattern)
	}

	// Add pagination
	offset := (page - 1) * pageSize
	query = query.Limit(uint64(pageSize)).Offset(uint64(offset))

	// Get total count
	countQuery := query.Column("COUNT(*) OVER()")
	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var communities []models.Community
	var total int64

	for rows.Next() {
		var community models.Community
		err := rows.Scan(
			&community.ID,
			&community.Name,
			&community.Abbreviation,
			&community.LeadID,
			&community.ProfilePhotoFileID,
			&community.CreatedAt,
			&community.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}
		communities = append(communities, community)
	}

	// Load profile photo file for each community if needed
	for i := range communities {
		if communities[i].ProfilePhotoFileID != nil {
			profilePhoto, err := r.getProfilePhoto(ctx, *communities[i].ProfilePhotoFileID)
			if err != nil {
				return nil, 0, fmt.Errorf("error getting profile photo for community %d: %w", communities[i].ID, err)
			}
			communities[i].ProfilePhoto = profilePhoto
		}
	}

	return communities, total, nil
}

// GetByID retrieves a community by ID
func (r *CommunityRepository) GetByID(ctx context.Context, id int64) (*models.Community, error) {
	query := squirrel.Select(
		"id", "name", "abbreviation", "lead_id", "profile_photo_file_id", "created_at", "updated_at",
	).
		From("communities").
		Where("id = ?", id).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var community models.Community
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&community.ID,
		&community.Name,
		&community.Abbreviation,
		&community.LeadID,
		&community.ProfilePhotoFileID,
		&community.CreatedAt,
		&community.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	// Get the profile photo for this community if it exists
	if community.ProfilePhotoFileID != nil {
		profilePhoto, err := r.getProfilePhoto(ctx, *community.ProfilePhotoFileID)
		if err != nil {
			return nil, fmt.Errorf("error getting profile photo: %w", err)
		}
		community.ProfilePhoto = profilePhoto
	}

	return &community, nil
}

// Create creates a new community
func (r *CommunityRepository) Create(ctx context.Context, community *models.Community) (int64, error) {
	query := squirrel.Insert("communities").
		Columns(
			"name", "abbreviation", "lead_id", "profile_photo_file_id",
		).
		Values(
			community.Name, community.Abbreviation, community.LeadID, community.ProfilePhotoFileID,
		).
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

// Update updates an existing community
func (r *CommunityRepository) Update(ctx context.Context, community *models.Community) error {
	query := squirrel.Update("communities").
		Set("name", community.Name).
		Set("abbreviation", community.Abbreviation).
		Set("lead_id", community.LeadID).
		Set("profile_photo_file_id", community.ProfilePhotoFileID).
		Set("updated_at", time.Now()).
		Where("id = ?", community.ID).
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

// Delete deletes a community
func (r *CommunityRepository) Delete(ctx context.Context, id int64) error {
	query := squirrel.Delete("communities").
		Where("id = ?", id).
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

// AddFileToCommunity adds a file to a community
func (r *CommunityRepository) AddFileToCommunity(ctx context.Context, communityID int64, fileID int64) error {
	query := squirrel.Insert("community_files").
		Columns("community_id", "file_id").
		Values(communityID, fileID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

// RemoveFileFromCommunity removes a file from a community
func (r *CommunityRepository) RemoveFileFromCommunity(ctx context.Context, communityID int64, fileID int64) error {
	query := squirrel.Delete("community_files").
		Where("community_id = ?", communityID).
		Where("file_id = ?", fileID).
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

// getProfilePhoto retrieves a single file by ID
func (r *CommunityRepository) getProfilePhoto(ctx context.Context, fileID int64) (*models.File, error) {
	query := squirrel.Select(
		"id", "file_name", "file_path", "file_url", "file_size", 
		"file_type", "resource_type", "resource_id", "uploaded_by", 
		"created_at", "updated_at",
	).
		From("files").
		Where("id = ?", fileID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var file models.File
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&file.ID,
		&file.FileName,
		&file.FilePath,
		&file.FileURL,
		&file.FileSize,
		&file.FileType,
		&file.ResourceType,
		&file.ResourceID,
		&file.UploadedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return &file, nil
}

// UpdateProfilePhoto updates the profile photo file ID for a community
func (r *CommunityRepository) UpdateProfilePhoto(ctx context.Context, communityID int64, fileID *int64) error {
	query := squirrel.Update("communities").
		Set("profile_photo_file_id", fileID).
		Set("updated_at", time.Now()).
		Where("id = ?", communityID).
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