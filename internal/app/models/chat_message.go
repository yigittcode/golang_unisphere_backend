package models

import "time"

// ChatMessageType represents the type of chat message
type ChatMessageType string

const (
	ChatMessageTypeText ChatMessageType = "TEXT"
	ChatMessageTypeFile ChatMessageType = "FILE"
)

// ChatMessage represents a message in a community chat
type ChatMessage struct {
	ID          int64           `json:"id" db:"id"`
	CommunityID int64           `json:"communityId" db:"community_id"`
	SenderID    int64           `json:"senderId" db:"sender_id"`
	MessageType ChatMessageType `json:"messageType" db:"message_type"`
	Content     string          `json:"content" db:"content"`
	FileID      *int64          `json:"fileId,omitempty" db:"file_id"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`

	// Related entities
	Sender *User `json:"sender,omitempty"`
	File   *File `json:"file,omitempty"`
}