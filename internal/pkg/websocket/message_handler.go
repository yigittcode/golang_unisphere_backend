package websocket

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
)

// MessageHandler processes WebSocket messages and persists them to the database
type MessageHandler struct {
	chatRepo    *repositories.ChatRepository
	userRepo    *repositories.UserRepository
	hub         *Hub
	logger      zerolog.Logger
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(
	chatRepo *repositories.ChatRepository,
	userRepo *repositories.UserRepository,
	hub *Hub,
	logger zerolog.Logger,
) *MessageHandler {
	return &MessageHandler{
		chatRepo: chatRepo,
		userRepo: userRepo,
		hub:      hub,
		logger:   logger,
	}
}

// Start begins processing messages from the hub
func (h *MessageHandler) Start() {
	go h.processMessages()
}

// processMessages listens for messages and saves them to the database
func (h *MessageHandler) processMessages() {
	// Create a new channel for listening to messages
	messageChan := make(chan *Message)
	
	// Register handler with the hub
	h.hub.AddMessageListener(messageChan)
	
	// Process messages
	for message := range messageChan {
		if message.Type == "text" {
			h.processTextMessage(message)
		} else if message.Type == "file" {
			// Skip file messages - they are handled directly by the service
			continue
		}
	}
}

// processTextMessage saves a text message to the database
func (h *MessageHandler) processTextMessage(message *Message) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Create a chat message model
	chatMessage := &models.ChatMessage{
		CommunityID: message.CommunityID,
		SenderID:    message.SenderID,
		MessageType: models.ChatMessageTypeText,
		Content:     message.Content,
	}
	
	// Save message to database
	messageID, err := h.chatRepo.Create(ctx, chatMessage)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int64("communityID", message.CommunityID).
			Int64("senderID", message.SenderID).
			Msg("Failed to save WebSocket message to database")
		return
	}
	
	// Update the message ID
	message.ID = messageID
	
	h.logger.Debug().
		Int64("messageID", messageID).
		Int64("communityID", message.CommunityID).
		Msg("WebSocket message saved to database")
}

// HandleIncomingMessage processes an incoming WebSocket message and saves it to the database
func (h *MessageHandler) HandleIncomingMessage(message *Message) {
	// Save to database based on message type
	if message.Type == "text" {
		go h.processTextMessage(message)
	}
	
	// Broadcast to clients
	h.hub.BroadcastToCommunity(message)
}