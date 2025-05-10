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

// ChatRepository handles database operations for chat messages
type ChatRepository struct {
	db *pgxpool.Pool
}

// NewChatRepository creates a new ChatRepository
func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

// Create inserts a new chat message into the database
func (r *ChatRepository) Create(ctx context.Context, message *models.ChatMessage) (int64, error) {
	query := `
		INSERT INTO chat_messages (
			community_id, sender_id, message_type, content, file_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		message.CommunityID,
		message.SenderID,
		message.MessageType,
		message.Content,
		message.FileID,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return 0, fmt.Errorf("error creating chat message: %w", err)
	}

	message.ID = id
	message.CreatedAt = createdAt
	message.UpdatedAt = updatedAt

	return id, nil
}

// GetByID retrieves a message by its ID
func (r *ChatRepository) GetByID(ctx context.Context, id int64) (*models.ChatMessage, error) {
	query := `
		SELECT 
			id, community_id, sender_id, message_type, content, file_id, created_at, updated_at
		FROM chat_messages
		WHERE id = $1
	`

	var message models.ChatMessage
	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.CommunityID,
		&message.SenderID,
		&message.MessageType,
		&message.Content,
		&message.FileID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chat message not found with ID %d", id)
		}
		return nil, fmt.Errorf("error retrieving chat message: %w", err)
	}

	return &message, nil
}

// GetByCommunityID retrieves messages for a specific community with filters
func (r *ChatRepository) GetByCommunityID(
	ctx context.Context,
	communityID int64,
	before *time.Time,
	after *time.Time,
	senderID *int64,
	limit int,
) ([]*models.ChatMessage, error) {
	// Base query
	queryBuilder := squirrel.Select(
		"cm.id", "cm.community_id", "cm.sender_id", "cm.message_type", 
		"cm.content", "cm.file_id", "cm.created_at", "cm.updated_at",
		"u.id as user_id", "u.first_name", "u.last_name", "u.email",
		"f.id as file_id", "f.file_name", "f.file_url", "f.file_type", "f.file_size",
	).
		From("chat_messages cm").
		LeftJoin("users u ON cm.sender_id = u.id").
		LeftJoin("files f ON cm.file_id = f.id").
		Where("cm.community_id = ?", communityID).
		OrderBy("cm.created_at DESC").
		Limit(uint64(limit)).
		PlaceholderFormat(squirrel.Dollar)

	// Add filters if provided
	if before != nil {
		queryBuilder = queryBuilder.Where("cm.created_at < ?", before)
	}

	if after != nil {
		queryBuilder = queryBuilder.Where("cm.created_at > ?", after)
	}

	if senderID != nil {
		queryBuilder = queryBuilder.Where("cm.sender_id = ?", *senderID)
	}

	// Build the SQL query
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	// Execute the query
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		var message models.ChatMessage
		var user models.User
		var file models.File
		var fileID, userID *int64
		var firstName, lastName, email, fileName, fileURL, fileType *string
		var fileSize *int64

		err := rows.Scan(
			&message.ID,
			&message.CommunityID,
			&message.SenderID,
			&message.MessageType,
			&message.Content,
			&message.FileID,
			&message.CreatedAt,
			&message.UpdatedAt,
			&userID,
			&firstName,
			&lastName,
			&email,
			&fileID,
			&fileName,
			&fileURL,
			&fileType,
			&fileSize,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning chat message row: %w", err)
		}

		// Add the sender if exists
		if userID != nil {
			user.ID = *userID
			if firstName != nil {
				user.FirstName = *firstName
			}
			if lastName != nil {
				user.LastName = *lastName
			}
			if email != nil {
				user.Email = *email
			}
			message.Sender = &user
		}

		// Add the file if exists
		if fileID != nil && message.FileID != nil {
			file.ID = *fileID
			if fileName != nil {
				file.FileName = *fileName
			}
			if fileURL != nil {
				file.FileURL = *fileURL
			}
			if fileType != nil {
				file.FileType = *fileType
			}
			if fileSize != nil {
				file.FileSize = *fileSize
			}
			message.File = &file
		}

		messages = append(messages, &message)
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat message rows: %w", err)
	}

	return messages, nil
}

// Delete removes a chat message
func (r *ChatRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM chat_messages WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting chat message: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no chat message found with ID %d", id)
	}

	return nil
}

// DeleteByCommunity removes all chat messages for a community
func (r *ChatRepository) DeleteByCommunity(ctx context.Context, communityID int64) error {
	query := `DELETE FROM chat_messages WHERE community_id = $1`

	_, err := r.db.Exec(ctx, query, communityID)
	if err != nil {
		return fmt.Errorf("error deleting chat messages for community: %w", err)
	}

	return nil
}