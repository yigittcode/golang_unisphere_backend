package dto

import (
	"time"

	"github.com/yigit/unisphere/internal/app/models"
)

// --- Request DTOs ---

// CreateChatMessageRequest represents data for creating a new chat message
type CreateChatMessageRequest struct {
	MessageType string `json:"messageType" form:"messageType" binding:"required,oneof=TEXT FILE"`
	Content     string `json:"content" form:"content" binding:"required_if=MessageType TEXT"`
}

// GetChatMessagesRequest represents filter parameters for retrieving chat messages
type GetChatMessagesRequest struct {
	Before  *time.Time `form:"before,omitempty"`
	After   *time.Time `form:"after,omitempty"`
	Limit   int        `form:"limit,default=50" binding:"min=1,max=100"`
	SenderID *int64    `form:"senderId,omitempty"`
}

// --- Response DTOs ---

// ChatMessageResponse represents a chat message with basic information
type ChatMessageResponse struct {
	ID          int64     `json:"id"`
	CommunityID int64     `json:"communityId"`
	SenderID    int64     `json:"senderId"`
	MessageType string    `json:"messageType"`
	Content     string    `json:"content"`
	FileID      *int64    `json:"fileId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
	// Sender information
	SenderName  string  `json:"senderName,omitempty"`
	
	// File information if available
	FileName     *string `json:"fileName,omitempty"`
	FileURL      *string `json:"fileUrl,omitempty"`
	FileType     *string `json:"fileType,omitempty"`
}

// ChatMessageDetailResponse extends ChatMessageResponse with full sender and file details
type ChatMessageDetailResponse struct {
	ChatMessageResponse
	Sender *UserBasicResponse `json:"sender,omitempty"`
	File   *ChatFileResponse  `json:"file,omitempty"`
}

// ChatFileResponse represents file details for chat messages
type ChatFileResponse struct {
	ID       int64  `json:"id"`
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
	FileType string `json:"fileType"`
	FileSize int64  `json:"fileSize"`
}

// ChatMessageListResponse represents a list of chat messages
type ChatMessageListResponse struct {
	Messages []ChatMessageResponse `json:"messages"`
}

// Transform a models.ChatMessage to ChatMessageResponse
func ToChatMessageResponse(message *models.ChatMessage) ChatMessageResponse {
	response := ChatMessageResponse{
		ID:          message.ID,
		CommunityID: message.CommunityID,
		SenderID:    message.SenderID,
		MessageType: string(message.MessageType),
		Content:     message.Content,
		FileID:      message.FileID,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}

	// Add sender name if sender is available
	if message.Sender != nil {
		response.SenderName = message.Sender.FirstName + " " + message.Sender.LastName
	}

	// Add file information if a file is attached
	if message.File != nil {
		fileName := message.File.FileName
		fileURL := message.File.FileURL
		fileType := message.File.FileType
		
		response.FileName = &fileName
		response.FileURL = &fileURL
		response.FileType = &fileType
	}

	return response
}

// Transform a models.ChatMessage to ChatMessageDetailResponse
func ToChatMessageDetailResponse(message *models.ChatMessage) ChatMessageDetailResponse {
	response := ChatMessageDetailResponse{
		ChatMessageResponse: ToChatMessageResponse(message),
	}

	// Add full sender details if available
	if message.Sender != nil {
		response.Sender = &UserBasicResponse{
			ID:        message.Sender.ID,
			FirstName: message.Sender.FirstName,
			LastName:  message.Sender.LastName,
			Email:     message.Sender.Email,
		}
	}

	// Add full file details if available
	if message.File != nil {
		response.File = &ChatFileResponse{
			ID:       message.File.ID,
			FileName: message.File.FileName,
			FileURL:  message.File.FileURL,
			FileType: message.File.FileType,
			FileSize: message.File.FileSize,
		}
	}

	return response
}