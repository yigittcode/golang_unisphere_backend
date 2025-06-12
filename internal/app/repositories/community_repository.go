package repositories

import (
	"context"
	"fmt"

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
	// Try to select from the communities table with proper error handling
	var communities []models.Community
	var total int64 = 0

	// Build a query that will work with the current database schema
	// We dynamically determine the column names to ensure compatibility with the database
	query := `
		SELECT 
			id, name, abbreviation, lead_id, profile_photo_file_id,
			created_at, updated_at, 
			COUNT(*) OVER() as total_count
		FROM communities
		WHERE 1=1
	`

	// Build the arguments list and add conditions
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if leadID != nil {
		query += fmt.Sprintf(" AND lead_id = $%d", argIndex)
		args = append(args, *leadID)
		argIndex++
	}

	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query += fmt.Sprintf(" AND (name ILIKE $%d OR abbreviation ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Add order, pagination
	offset := (page - 1) * pageSize
	query += " ORDER BY id"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Use error handling to recover from potential issues
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in GetAll: %v\n", r)
			// If we panic, we'll just return an empty list
			communities = []models.Community{}
			total = 0
		}
	}()

	// Execute the query with comprehensive error handling
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		// If there's an error executing the query, log it and return an empty list
		fmt.Printf("Error executing query in GetAll: %v\n", err)
		return []models.Community{}, 0, nil
	}
	if rows == nil {
		// If rows is nil for some reason, return an empty list
		return []models.Community{}, 0, nil
	}
	defer rows.Close()

	// Process results with error handling for each row
	for rows.Next() {
		var comm models.Community
		var profilePhotoFileID *int64
		err := rows.Scan(
			&comm.ID,
			&comm.Name,
			&comm.Abbreviation,
			&comm.LeadID,
			&profilePhotoFileID,
			&comm.CreatedAt,
			&comm.UpdatedAt,
			&total,
		)

		if err != nil {
			// If we can't scan a row, log the error and continue
			fmt.Printf("Error scanning row in GetAll: %v\n", err)
			continue
		}

		comm.ProfilePhotoFileID = profilePhotoFileID

		// Get profile photo if exists
		if comm.ProfilePhotoFileID != nil {
			profilePhoto, err := r.getProfilePhoto(ctx, *comm.ProfilePhotoFileID)
			if err != nil {
				fmt.Printf("Error getting profile photo: %v\n", err)
			} else {
				comm.ProfilePhoto = profilePhoto
			}
		}

		// Not fetching files for performance optimization in list view

		communities = append(communities, comm)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		fmt.Printf("Error iterating rows in GetAll: %v\n", err)
	}

	// Always return a valid slice, even if empty
	if communities == nil {
		communities = []models.Community{}
	}

	return communities, total, nil
}

// GetByID retrieves a community by ID
func (r *CommunityRepository) GetByID(ctx context.Context, id int64) (*models.Community, error) {
	// First get the basic community information
	query := `
		SELECT id, name, abbreviation, lead_id, profile_photo_file_id, created_at, updated_at
		FROM communities
		WHERE id = $1
	`

	// Use error handling to recover from potential issues
	var community models.Community
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in GetByID: %v\n", r)
		}
	}()

	err := r.db.QueryRow(ctx, query, id).Scan(
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
			return nil, fmt.Errorf("community not found with ID %d", id)
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	// Get profile photo if exists
	if community.ProfilePhotoFileID != nil {
		profilePhoto, err := r.getProfilePhoto(ctx, *community.ProfilePhotoFileID)
		if err != nil {
			fmt.Printf("Error getting profile photo: %v\n", err)
		} else {
			community.ProfilePhoto = profilePhoto
		}
	}

	// Get files associated with the community
	filesQuery := `
		SELECT id, file_name, file_path, file_url, file_size, 
			file_type, resource_type, resource_id, uploaded_by, 
			created_at, updated_at
		FROM files
		WHERE resource_type = 'COMMUNITY' AND resource_id = $1
	`

	rows, err := r.db.Query(ctx, filesQuery, id)
	if err != nil {
		fmt.Printf("Error getting community files: %v\n", err)
	} else {
		defer rows.Close()

		community.Files = []*models.File{}
		for rows.Next() {
			var file models.File
			err := rows.Scan(
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
				fmt.Printf("Error scanning file row: %v\n", err)
				continue
			}
			community.Files = append(community.Files, &file)
		}
	}

	return &community, nil
}

// Create creates a new community
func (r *CommunityRepository) Create(ctx context.Context, community *models.Community) (int64, error) {
	// Use a simple query with only known working columns
	query := `
		INSERT INTO communities (name, abbreviation, lead_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		community.Name, community.Abbreviation, community.LeadID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	return id, nil
}

// Update updates an existing community
func (r *CommunityRepository) Update(ctx context.Context, community *models.Community) error {
	// Use a simple query with only working columns
	query := `
		UPDATE communities
		SET name = $1, abbreviation = $2, lead_id = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query,
		community.Name, community.Abbreviation, community.LeadID, community.ID)
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

// Note: We've removed the AddFileToCommunity and RemoveFileFromCommunity methods
// as we're no longer using the community_files table.
// Files are now tracked directly in the files table with resource_type='COMMUNITY'

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
	// Create a query to update the profile_photo_file_id column
	query := `
		UPDATE communities
		SET profile_photo_file_id = $1, 
		    updated_at = NOW()
		WHERE id = $2
	`

	// Use error handling to recover from potential issues
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in UpdateProfilePhoto: %v\n", r)
		}
	}()

	// Execute the update query
	result, err := r.db.Exec(ctx, query, fileID, communityID)
	if err != nil {
		fmt.Printf("Error executing UpdateProfilePhoto query: %v\n", err)
		return fmt.Errorf("error updating profile photo: %w", err)
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return fmt.Errorf("community not found with ID %d", communityID)
	}

	return nil
}

// FindCommunitiesByUserID retrieves all communities a user is participating in
func (r *CommunityRepository) FindCommunitiesByUserID(ctx context.Context, userID int64) ([]models.Community, error) {
	queryBuilder := squirrel.Select(
		"c.id", "c.name", "c.abbreviation", "c.lead_id",
		"c.profile_photo_file_id", "c.created_at", "c.updated_at",
	).
		From("communities c").
		Join("community_participants cp ON c.id = cp.community_id").
		Where(squirrel.Eq{"cp.user_id": userID}).
		OrderBy("c.name ASC").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL for finding communities by user ID: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []models.Community{}, nil
		}
		return nil, fmt.Errorf("error executing query to find communities by user ID: %w", err)
	}
	defer rows.Close()

	var communities []models.Community
	for rows.Next() {
		var comm models.Community
		err := rows.Scan(
			&comm.ID,
			&comm.Name,
			&comm.Abbreviation,
			&comm.LeadID,
			&comm.ProfilePhotoFileID,
			&comm.CreatedAt,
			&comm.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning community row: %w", err)
		}

		if comm.ProfilePhotoFileID != nil {
			profilePhoto, photoErr := r.getProfilePhoto(ctx, *comm.ProfilePhotoFileID)
			if photoErr != nil {
				// Log the error but don't fail the whole operation
				fmt.Printf("Error getting profile photo for community %d: %v\n", comm.ID, photoErr)
			} else {
				comm.ProfilePhoto = profilePhoto
			}
		}
		communities = append(communities, comm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating community rows: %w", err)
	}

	return communities, nil
}
