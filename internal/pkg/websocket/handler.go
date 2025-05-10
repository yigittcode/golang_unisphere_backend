package websocket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// Handler for WebSocket connections
type Handler struct {
	hub                     *Hub
	communityParticipantRepo *repositories.CommunityParticipantRepository
	logger                  zerolog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	hub *Hub,
	communityParticipantRepo *repositories.CommunityParticipantRepository,
	logger zerolog.Logger,
) *Handler {
	return &Handler{
		hub:                     hub,
		communityParticipantRepo: communityParticipantRepo,
		logger:                  logger,
	}
}

// HandleConnection godoc
// @Summary Establish a WebSocket connection for real-time chat
// @Description Upgrades HTTP connection to a WebSocket connection for real-time chat messaging
// @Tags chat, websocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 101 {string} string "Switching Protocols to WebSocket"
// @Failure 400 {object} gin.H "Invalid community ID"
// @Failure 401 {object} gin.H "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} gin.H "Forbidden: User is not a participant in the community"
// @Failure 500 {object} gin.H "Internal Server Error"
// @Router /communities/{id}/chat/ws [get]
func (h *Handler) HandleConnection(c *gin.Context) {
	// Get community ID from path
	communityIDStr := c.Param("id")
	communityID, err := strconv.ParseInt(communityIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid community ID",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	// Convert to int64
	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Check if user is a participant in the community
	isParticipant, err := h.communityParticipantRepo.IsUserParticipant(c, communityID, userID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int64("communityID", communityID).
			Int64("userID", userID).
			Msg("Failed to check if user is participant")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check participant status",
		})
		return
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{
			"error": apperrors.NewForbiddenError("User is not a participant in this community").Error(),
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int64("communityID", communityID).
			Int64("userID", userID).
			Msg("Failed to upgrade connection to WebSocket")
		return
	}

	// Create a new client and register it with the hub
	client := &Client{
		hub:         h.hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      userID,
		communityID: communityID,
		logger:      h.logger,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	h.logger.Info().
		Int64("communityID", communityID).
		Int64("userID", userID).
		Str("remoteAddr", conn.RemoteAddr().String()).
		Msg("WebSocket connection established")
}