package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// ChatController handles chat message operations
type ChatController struct {
	chatService services.ChatService
}

// NewChatController creates a new ChatController
func NewChatController(chatService services.ChatService) *ChatController {
	return &ChatController{
		chatService: chatService,
	}
}

// GetChatMessages godoc
// @Summary Get community chat messages
// @Description Retrieve chat messages for a specific community with pagination and filters
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param before query string false "Get messages before this timestamp (RFC3339 format)"
// @Param after query string false "Get messages after this timestamp (RFC3339 format)"
// @Param limit query int false "Maximum number of messages to retrieve (default: 50)" default(50)
// @Param senderId query int false "Filter messages by sender ID"
// @Success 200 {object} dto.APIResponse{data=[]dto.ChatMessageResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User is not a participant in the community"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail} "Community not found"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/chat [get]
func (c *ChatController) GetChatMessages(ctx *gin.Context) {
	fmt.Println("********* GetChatMessages STARTED *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Parse filters from query parameters
	filter := &dto.GetChatMessagesRequest{
		Limit: 50, // Default limit
	}

	// Parse limit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	// Parse before time
	if beforeStr := ctx.Query("before"); beforeStr != "" {
		beforeTime, err := time.Parse(time.RFC3339, beforeStr)
		if err == nil {
			filter.Before = &beforeTime
		}
	}

	// Parse after time
	if afterStr := ctx.Query("after"); afterStr != "" {
		afterTime, err := time.Parse(time.RFC3339, afterStr)
		if err == nil {
			filter.After = &afterTime
		}
	}

	// Parse sender ID
	if senderIDStr := ctx.Query("senderId"); senderIDStr != "" {
		senderID, err := strconv.ParseInt(senderIDStr, 10, 64)
		if err == nil && senderID > 0 {
			filter.SenderID = &senderID
		}
	}

	// Get messages from service
	messages, err := c.chatService.GetChatMessages(ctx, communityID, filter)
	if err != nil {
		fmt.Printf("Error getting chat messages: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community not found")))
		case errors.Is(err, apperrors.ErrPermissionDenied) || strings.Contains(err.Error(), "not a participant"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "User is not a participant in this community")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to retrieve chat messages").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* GetChatMessages SUCCESSFUL *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(messages))
}

// GetChatMessageByID godoc
// @Summary Get chat message by ID
// @Description Retrieve a specific chat message by its ID
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param messageId path int true "Message ID"
// @Success 200 {object} dto.APIResponse{data=dto.ChatMessageDetailResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User is not a participant in the community"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail} "Message not found"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/chat/{messageId} [get]
func (c *ChatController) GetChatMessageByID(ctx *gin.Context) {
	fmt.Println("********* GetChatMessageByID STARTED *********")

	// Parse message ID from path
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid message ID")))
		return
	}

	// Get message from service
	message, err := c.chatService.GetChatMessageByID(ctx, messageID)
	if err != nil {
		fmt.Printf("Error getting chat message: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Chat message not found")))
		case errors.Is(err, apperrors.ErrPermissionDenied) || strings.Contains(err.Error(), "not a participant"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "User is not a participant in this community")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to retrieve chat message").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* GetChatMessageByID SUCCESSFUL *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(message))
}

// SendTextMessage godoc
// @Summary Send a text message to a community chat
// @Description Send a new text message to a specific community's chat
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param message body dto.CreateChatMessageRequest true "Message details"
// @Success 201 {object} dto.APIResponse{data=dto.ChatMessageResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User is not a participant in the community"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail} "Community not found"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/chat/text [post]
func (c *ChatController) SendTextMessage(ctx *gin.Context) {
	fmt.Println("********* SendTextMessage STARTED *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Parse request message
	var req dto.CreateChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format").WithDetails(err.Error())))
		return
	}

	// Set message type to TEXT
	req.MessageType = string(models.ChatMessageTypeText)

	// Send message
	message, err := c.chatService.SendTextMessage(ctx, communityID, &req)
	if err != nil {
		fmt.Printf("Error sending text message: %v\n", err)
		middleware.HandleAPIError(ctx, err)
		return
	}

	fmt.Println("********* SendTextMessage SUCCESSFUL *********")
	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(message))
}

// SendFileMessage godoc
// @Summary Send a file message to a community chat
// @Description Send a new file message to a specific community's chat
// @Tags chat
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param content formData string false "Optional message content to accompany the file"
// @Param file formData file true "File to upload (PDF or image)"
// @Success 201 {object} dto.APIResponse{data=dto.ChatMessageResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User is not a participant in the community"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail} "Community not found"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/chat/file [post]
func (c *ChatController) SendFileMessage(ctx *gin.Context) {
	fmt.Println("********* SendFileMessage STARTED *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "No file provided").WithDetails(err.Error())))
		return
	}

	// Get content from form
	content := ctx.PostForm("content")

	// Create message request
	req := &dto.CreateChatMessageRequest{
		MessageType: string(models.ChatMessageTypeFile),
		Content:     content,
	}

	// Send file message
	message, err := c.chatService.SendFileMessage(ctx, communityID, req, file)
	if err != nil {
		fmt.Printf("Error sending file message: %v\n", err)
		middleware.HandleAPIError(ctx, err)
		return
	}

	fmt.Println("********* SendFileMessage SUCCESSFUL *********")
	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(message))
}

// DeleteChatMessage godoc
// @Summary Delete a chat message
// @Description Delete a specific chat message (only available to message sender or community lead)
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param messageId path int true "Message ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse}
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 403 {object} dto.APIResponse{error=dto.ErrorDetail} "Forbidden: User is not the message sender or community lead"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail} "Message not found"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/chat/{messageId} [delete]
func (c *ChatController) DeleteChatMessage(ctx *gin.Context) {
	fmt.Println("********* DeleteChatMessage STARTED *********")

	// Parse message ID from path
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid message ID")))
		return
	}

	// Delete message
	err = c.chatService.DeleteMessage(ctx, messageID)
	if err != nil {
		fmt.Printf("Error deleting chat message: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Chat message not found")))
		case errors.Is(err, apperrors.ErrPermissionDenied) || strings.Contains(err.Error(), "only the message sender"):
			ctx.JSON(http.StatusForbidden, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeForbidden, "Only the message sender or community lead can delete messages")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete chat message").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* DeleteChatMessage SUCCESSFUL *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "Chat message deleted successfully"}))
}