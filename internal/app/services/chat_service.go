package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/websocket"
)

// ChatService defines the interface for chat operations
type ChatService interface {
	GetChatMessages(ctx context.Context, communityID int64, filter *dto.GetChatMessagesRequest) ([]dto.ChatMessageResponse, error)
	GetChatMessageByID(ctx context.Context, messageID int64) (*dto.ChatMessageDetailResponse, error)
	SendTextMessage(ctx context.Context, communityID int64, message *dto.CreateChatMessageRequest) (*dto.ChatMessageResponse, error)
	SendFileMessage(ctx context.Context, communityID int64, message *dto.CreateChatMessageRequest, file *multipart.FileHeader) (*dto.ChatMessageResponse, error)
	DeleteMessage(ctx context.Context, messageID int64) error
}

// chatServiceImpl implements ChatService
type chatServiceImpl struct {
	chatRepo                 *repositories.ChatRepository
	communityRepo            *repositories.CommunityRepository
	communityParticipantRepo *repositories.CommunityParticipantRepository
	userRepo                 *repositories.UserRepository
	fileRepo                 *repositories.FileRepository
	fileStorage              *filestorage.LocalStorage
	wsHub                    *websocket.Hub // WebSocket hub for real-time messaging
	logger                   zerolog.Logger
}

// NewChatService creates a new ChatService
func NewChatService(
	chatRepo *repositories.ChatRepository,
	communityRepo *repositories.CommunityRepository,
	communityParticipantRepo *repositories.CommunityParticipantRepository,
	userRepo *repositories.UserRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	wsHub *websocket.Hub,
	logger zerolog.Logger,
) ChatService {
	return &chatServiceImpl{
		chatRepo:                 chatRepo,
		communityRepo:            communityRepo,
		communityParticipantRepo: communityParticipantRepo,
		userRepo:                 userRepo,
		fileRepo:                 fileRepo,
		fileStorage:              fileStorage,
		wsHub:                    wsHub,
		logger:                   logger,
	}
}

// GetChatMessages retrieves chat messages for a community
func (s *chatServiceImpl) GetChatMessages(
	ctx context.Context,
	communityID int64,
	filter *dto.GetChatMessagesRequest,
) ([]dto.ChatMessageResponse, error) {
	s.logger.Debug().
		Int64("communityID", communityID).
		Interface("filter", filter).
		Msg("Retrieving chat messages")

	// Check if community exists
	community, err := s.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to get community")
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}
	
	if community == nil {
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}

	// Get user ID from context for authorization
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is a participant in the community
	isParticipant, err := s.communityParticipantRepo.IsUserParticipant(ctx, communityID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Int64("userID", userID).
			Msg("Failed to check if user is participant")
		return nil, fmt.Errorf("error checking participant status: %w", err)
	}

	if !isParticipant {
		return nil, apperrors.NewForbiddenError("User is not a participant in this community")
	}

	// Retrieve messages
	messages, err := s.chatRepo.GetByCommunityID(
		ctx,
		communityID,
		filter.Before,
		filter.After,
		filter.SenderID,
		filter.Limit,
	)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to retrieve chat messages")
		return nil, fmt.Errorf("error retrieving chat messages: %w", err)
	}

	// Convert to DTOs
	var responseMessages []dto.ChatMessageResponse
	for _, message := range messages {
		responseMessages = append(responseMessages, dto.ToChatMessageResponse(message))
	}

	return responseMessages, nil
}

// GetChatMessageByID retrieves a specific chat message
func (s *chatServiceImpl) GetChatMessageByID(
	ctx context.Context,
	messageID int64,
) (*dto.ChatMessageDetailResponse, error) {
	s.logger.Debug().
		Int64("messageID", messageID).
		Msg("Retrieving chat message by ID")

	// Get message from repository
	message, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("messageID", messageID).
			Msg("Failed to get chat message")
		return nil, apperrors.NewResourceNotFoundError("Chat message not found")
	}

	// Get user ID from context for authorization
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is a participant in the community
	isParticipant, err := s.communityParticipantRepo.IsUserParticipant(ctx, message.CommunityID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", message.CommunityID).
			Int64("userID", userID).
			Msg("Failed to check if user is participant")
		return nil, fmt.Errorf("error checking participant status: %w", err)
	}

	if !isParticipant {
		return nil, apperrors.NewForbiddenError("User is not a participant in this community")
	}

	// Populate sender information if not already loaded
	if message.Sender == nil && message.SenderID > 0 {
		sender, err := s.userRepo.FindByID(ctx, message.SenderID)
		if err == nil && sender != nil {
			message.Sender = sender
		}
	}

	// Populate file information if not already loaded
	if message.File == nil && message.FileID != nil {
		file, err := s.fileRepo.GetByID(ctx, *message.FileID)
		if err == nil && file != nil {
			message.File = file
		}
	}

	response := dto.ToChatMessageDetailResponse(message)
	return &response, nil
}

// SendTextMessage sends a new text message to a community chat
func (s *chatServiceImpl) SendTextMessage(
	ctx context.Context,
	communityID int64,
	messageReq *dto.CreateChatMessageRequest,
) (*dto.ChatMessageResponse, error) {
	s.logger.Debug().
		Int64("communityID", communityID).
		Str("messageType", messageReq.MessageType).
		Msg("Sending text message to community chat")

	// Check if message type is TEXT
	if messageReq.MessageType != string(models.ChatMessageTypeText) {
		return nil, apperrors.NewBadRequestError("Message type must be TEXT for text messages")
	}

	// Check if content is provided
	if messageReq.Content == "" {
		return nil, apperrors.NewBadRequestError("Content is required for text messages")
	}

	// Check if community exists
	community, err := s.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to get community")
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}
	
	if community == nil {
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is a participant in the community
	isParticipant, err := s.communityParticipantRepo.IsUserParticipant(ctx, communityID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Int64("userID", userID).
			Msg("Failed to check if user is participant")
		return nil, fmt.Errorf("error checking participant status: %w", err)
	}

	if !isParticipant {
		return nil, apperrors.NewForbiddenError("User is not a participant in this community")
	}

	// Create message model
	message := &models.ChatMessage{
		CommunityID: communityID,
		SenderID:    userID,
		MessageType: models.ChatMessageTypeText,
		Content:     messageReq.Content,
	}

	// Save message to database
	_, err = s.chatRepo.Create(ctx, message)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to create chat message")
		return nil, fmt.Errorf("error creating chat message: %w", err)
	}

	// Get user information for response
	user, err := s.userRepo.FindByID(ctx, userID)
	if err == nil && user != nil {
		message.Sender = user
	}

	// Broadcast the message through WebSocket if the hub is available
	if s.wsHub != nil {
		// Create WebSocket message
		wsMessage := &websocket.Message{
			Type:        "text",
			CommunityID: communityID,
			SenderID:    userID,
			Content:     messageReq.Content,
			Timestamp:   message.CreatedAt,
			ID:          message.ID,
		}

		// Broadcast to all connected clients in the community
		s.wsHub.BroadcastToCommunity(wsMessage)
		s.logger.Debug().
			Int64("communityID", communityID).
			Int64("messageID", message.ID).
			Msg("Text message broadcasted via WebSocket")
	}

	response := dto.ToChatMessageResponse(message)
	return &response, nil
}

// SendFileMessage sends a new file message to a community chat
func (s *chatServiceImpl) SendFileMessage(
	ctx context.Context,
	communityID int64,
	messageReq *dto.CreateChatMessageRequest,
	fileHeader *multipart.FileHeader,
) (*dto.ChatMessageResponse, error) {
	s.logger.Debug().
		Int64("communityID", communityID).
		Str("fileName", fileHeader.Filename).
		Msg("Sending file message to community chat")

	// Check if message type is FILE
	if messageReq.MessageType != string(models.ChatMessageTypeFile) {
		return nil, apperrors.NewBadRequestError("Message type must be FILE for file messages")
	}

	// Check if community exists
	community, err := s.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to get community")
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}
	
	if community == nil {
		return nil, apperrors.NewResourceNotFoundError("Community not found")
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Check if user is a participant in the community
	isParticipant, err := s.communityParticipantRepo.IsUserParticipant(ctx, communityID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Int64("userID", userID).
			Msg("Failed to check if user is participant")
		return nil, fmt.Errorf("error checking participant status: %w", err)
	}

	if !isParticipant {
		return nil, apperrors.NewForbiddenError("User is not a participant in this community")
	}

	// Upload the file
	uploadedFile, err := s.uploadFile(ctx, fileHeader, models.FileTypeChatMessage, communityID, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("filename", fileHeader.Filename).
			Int64("communityID", communityID).
			Msg("Failed to upload file for chat message")
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create content for file message
	content := ""
	if messageReq.Content != "" {
		content = messageReq.Content
	} else {
		content = "File: " + fileHeader.Filename
	}

	// Create message model
	message := &models.ChatMessage{
		CommunityID: communityID,
		SenderID:    userID,
		MessageType: models.ChatMessageTypeFile,
		Content:     content,
		FileID:      &uploadedFile.ID,
		File:        uploadedFile,
	}

	// Save message to database
	_, err = s.chatRepo.Create(ctx, message)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("communityID", communityID).
			Msg("Failed to create chat message")
		
		// Clean up - delete the file if we couldn't create the message
		_ = s.fileStorage.DeleteFile(uploadedFile.FilePath)
		_ = s.fileRepo.Delete(ctx, uploadedFile.ID)
		
		return nil, fmt.Errorf("error creating chat message: %w", err)
	}

	// Get user information for response
	user, err := s.userRepo.FindByID(ctx, userID)
	if err == nil && user != nil {
		message.Sender = user
	}

	// Broadcast the message through WebSocket if the hub is available
	if s.wsHub != nil {
		// Create WebSocket message
		wsMessage := &websocket.Message{
			Type:        "file",
			CommunityID: communityID,
			SenderID:    userID,
			Content:     content,
			FileURL:     uploadedFile.FileURL,
			FileID:      uploadedFile.ID,
			Timestamp:   message.CreatedAt,
			ID:          message.ID,
		}

		// Broadcast to all connected clients in the community
		s.wsHub.BroadcastToCommunity(wsMessage)
		s.logger.Debug().
			Int64("communityID", communityID).
			Int64("messageID", message.ID).
			Str("fileURL", uploadedFile.FileURL).
			Msg("File message broadcasted via WebSocket")
	}

	response := dto.ToChatMessageResponse(message)
	return &response, nil
}

// DeleteMessage deletes a chat message
func (s *chatServiceImpl) DeleteMessage(
	ctx context.Context,
	messageID int64,
) error {
	s.logger.Debug().
		Int64("messageID", messageID).
		Msg("Deleting chat message")

	// Get message from repository
	message, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("messageID", messageID).
			Msg("Failed to get chat message")
		return apperrors.NewResourceNotFoundError("Chat message not found")
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(int64)
	if !ok {
		s.logger.Error().Msg("User ID not found in context")
		return fmt.Errorf("user ID not found in context")
	}

	// Check if user is the sender or community lead
	community, err := s.communityRepo.GetByID(ctx, message.CommunityID)
	if err != nil || community == nil {
		s.logger.Error().Err(err).
			Int64("communityID", message.CommunityID).
			Msg("Failed to get community")
		return apperrors.NewResourceNotFoundError("Community not found")
	}

	if message.SenderID != userID && community.LeadID != userID {
		return apperrors.NewForbiddenError("Only the message sender or community lead can delete messages")
	}

	// Delete message
	err = s.chatRepo.Delete(ctx, messageID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("messageID", messageID).
			Msg("Failed to delete chat message")
		return fmt.Errorf("error deleting chat message: %w", err)
	}

	// Delete file if exists
	if message.FileID != nil {
		file, err := s.fileRepo.GetByID(ctx, *message.FileID)
		if err == nil && file != nil {
			// Delete physical file
			_ = s.fileStorage.DeleteFile(file.FilePath)
			// Delete file record
			_ = s.fileRepo.Delete(ctx, file.ID)
		}
	}

	// Broadcast message deletion if the hub is available
	if s.wsHub != nil {
		// Create WebSocket message for deletion notification
		wsMessage := &websocket.Message{
			Type:        "delete",
			CommunityID: message.CommunityID,
			SenderID:    userID,
			ID:          messageID,
			Timestamp:   message.CreatedAt,
		}

		// Broadcast to all connected clients in the community
		s.wsHub.BroadcastToCommunity(wsMessage)
		s.logger.Debug().
			Int64("communityID", message.CommunityID).
			Int64("messageID", messageID).
			Msg("Message deletion broadcasted via WebSocket")
	}

	return nil
}

// Helper method to upload a file
func (s *chatServiceImpl) uploadFile(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
	resourceType models.FileType,
	resourceID int64,
	userID int64,
) (*models.File, error) {
	// Open the file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer src.Close()

	// Generate a storage path based on resource type and ID
	subPath := fmt.Sprintf("%s_%d", resourceType, resourceID)

	// Upload to storage with the original FileHeader
	fileURL, err := s.fileStorage.SaveFileWithPath(fileHeader, subPath)
	if err != nil {
		return nil, fmt.Errorf("error uploading file: %w", err)
	}

	// Extract relative path from URL
	relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
	relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")

	// Create file model
	file := &models.File{
		FileName:     fileHeader.Filename,
		FilePath:     relativeFilePath,
		FileURL:      fileURL,
		FileSize:     fileHeader.Size,
		FileType:     fileHeader.Header.Get("Content-Type"),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UploadedBy:   userID,
	}

	// Save file metadata to database
	fileID, err := s.fileRepo.Create(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("error saving file metadata: %w", err)
	}
	file.ID = fileID

	return file, nil
}