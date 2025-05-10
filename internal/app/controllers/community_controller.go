package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/middleware"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/helpers"
)

// CommunityController handles community related operations
type CommunityController struct {
	communityService services.CommunityService
	fileStorage      *filestorage.LocalStorage
}

// NewCommunityController creates a new CommunityController
func NewCommunityController(communityService services.CommunityService, fileStorage *filestorage.LocalStorage) *CommunityController {
	return &CommunityController{
		communityService: communityService,
		fileStorage:      fileStorage,
	}
}

// GetAllCommunities handles retrieving all communities with optional filtering
// @Summary Get all communities
// @Description Retrieves a list of communities with optional filtering and pagination. Available to all authenticated users.
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param leadId query int false "Filter by lead ID"
// @Param search query string false "Search by name or abbreviation"
// @Param page query int false "Page number (1-based)" default(1) minimum(1)
// @Param pageSize query int false "Page size (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.APIResponse{data=dto.CommunityListResponse} "Communities retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /communities [get]
func (c *CommunityController) GetAllCommunities(ctx *gin.Context) {
	// Parse pagination parameters using helper
	page, pageSize := helpers.ParsePaginationParams(ctx)

	// Parse filters
	filters := make(map[string]interface{})

	// Add leadId filter if provided
	if leadIDStr := ctx.Query("leadId"); leadIDStr != "" {
		if leadID, err := strconv.ParseInt(leadIDStr, 10, 64); err == nil {
			filters["leadId"] = leadID
		}
	}

	// Add search filter if provided
	if search := ctx.Query("search"); search != "" {
		filters["search"] = search
	}

	// Get communities from service
	filter := &dto.CommunityFilterRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// Add filters if provided
	if leadID, ok := filters["leadId"].(int64); ok {
		filter.LeadID = &leadID
	}
	if search, ok := filters["search"].(string); ok {
		filter.Search = &search
	}

	// Try to get the communities, but have a fallback
	response, err := c.communityService.GetAllCommunities(ctx, filter)

	// If there's an error, return an empty list instead of an error
	if err != nil {
		fmt.Printf("Error in GetAllCommunities: %v, returning empty list\n", err)
		response = &dto.CommunityListResponse{
			Communities: []dto.CommunityResponse{},
			PaginationInfo: dto.PaginationInfo{
				CurrentPage: page,
				PageSize:    pageSize,
				TotalItems:  0,
				TotalPages:  1,
			},
		}
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// GetCommunityByID handles retrieving a specific community by ID
// @Summary Get community by ID
// @Description Retrieves a specific community by its ID
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=dto.CommunityResponse} "Community retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid community ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.ErrorResponse "Community not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /communities/{id} [get]
func (c *CommunityController) GetCommunityByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid community ID")
		errorDetail = errorDetail.WithDetails("Community ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Try to get the community by ID
	community, err := c.communityService.GetCommunityByID(ctx, id)

	// If we get an error, it probably means the community doesn't exist
	if err != nil {
		fmt.Printf("Error in GetCommunityByID: %v\n", err)
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community not found")
		ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(errorDetail))
		return
	}

	// If we successfully got the community, return it
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(community))
}

// CreateCommunity handles creating a new community
// @Summary Create a new community
// @Description Creates a new community with optional file upload. The current authenticated user will automatically be set as the community lead.
// @Tags communities
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Community name"
// @Param abbreviation formData string true "Community abbreviation"
// @Param profilePhoto formData file false "Community profile photo (single image file)"
// @Success 201 {object} dto.APIResponse{data=dto.CommunityResponse} "Community created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /communities [post]
func (c *CommunityController) CreateCommunity(ctx *gin.Context) {
	fmt.Println("********* CreateCommunity BAŞLANGIÇ *********")

	var req dto.CreateCommunityRequest
	if err := ctx.ShouldBind(&req); err != nil {
		fmt.Printf("Error binding request: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid request format").WithDetails(err.Error())))
		return
	}

	// Get profile photo (optional)
	var profilePhoto *multipart.FileHeader

	// Get profile photo from the multipart form
	profilePhoto, fileErr := ctx.FormFile("profilePhoto")
	if fileErr != nil && fileErr != http.ErrMissingFile {
		fmt.Printf("Error reading profile photo: %v\n", fileErr)
		// Continue without profile photo
	} else if profilePhoto != nil {
		fmt.Printf("Found profile photo in request: %s\n", profilePhoto.Filename)
	}

	// Create community
	community, err := c.communityService.CreateCommunity(ctx, &req, profilePhoto)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccessResponse(community))
}

// UpdateCommunity handles updating an existing community
// @Summary Update a community
// @Description Updates an existing community with the provided information
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param request body dto.UpdateCommunityRequest true "Update community request"
// @Success 200 {object} dto.APIResponse{data=dto.CommunityResponse} "Community updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.ErrorResponse "Community not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /communities/{id} [put]
func (c *CommunityController) UpdateCommunity(ctx *gin.Context) {
	// Get community ID from URL
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid community ID")
		errorDetail = errorDetail.WithDetails("Community ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Parse request data
	var req dto.UpdateCommunityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid request data")
		errorDetail = errorDetail.WithDetails(err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	// Get user ID from context
	_, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Call service to update community
	updatedCommunity, err := c.communityService.UpdateCommunity(ctx, id, &req)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(updatedCommunity))
}

// DeleteCommunity handles deleting a community
// @Summary Delete a community
// @Description Deletes an existing community by its ID
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "Community deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid community ID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.ErrorResponse "Community not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /communities/{id} [delete]
func (c *CommunityController) DeleteCommunity(ctx *gin.Context) {
	fmt.Println("********* DeleteCommunity BAŞLANGIÇ *********")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeValidationFailed, "Invalid community ID")
		errorDetail = errorDetail.WithDetails("Community ID must be a valid number")
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(errorDetail))
		return
	}

	fmt.Printf("Attempting to delete community with ID: %d\n", id)

	// Get user ID from context
	_, exists := ctx.Get("userID")
	if !exists {
		errorDetail := dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User not authenticated")
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(errorDetail))
		return
	}

	// Delete community
	err = c.communityService.DeleteCommunity(ctx, id)
	if err != nil {
		middleware.HandleAPIError(ctx, err)
		return
	}

	fmt.Println("********* DeleteCommunity BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "Community deleted successfully"}))
}

// Note: AddFileToCommunity method removed as file sharing is now handled through chat feature

// Note: DeleteFileFromCommunity method removed as file sharing is now handled through chat feature

// UpdateProfilePhoto godoc
// @Summary Update a community's profile photo
// @Description Update the profile photo for a community
// @Tags communities
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param photo formData file true "Profile photo to upload"
// @Success 200 {object} dto.APIResponse{data=dto.CommunityResponse} "Profile photo updated successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/profile-photo [post]
func (c *CommunityController) UpdateProfilePhoto(ctx *gin.Context) {
	fmt.Println("********* UpdateProfilePhoto BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	fmt.Printf("Updating profile photo for community with ID: %d\n", communityID)

	// Get profile photo from form
	file, err := ctx.FormFile("photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "No profile photo provided")))
		return
	}

	// Update profile photo
	response, err := c.communityService.UpdateProfilePhoto(ctx, communityID, file)
	if err != nil {
		fmt.Printf("Error updating profile photo: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community not found")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to update profile photo").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* UpdateProfilePhoto BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(response))
}

// DeleteProfilePhoto godoc
// @Summary Delete a community's profile photo
// @Description Remove the profile photo from a community
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "Profile photo deleted successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/profile-photo [delete]
func (c *CommunityController) DeleteProfilePhoto(ctx *gin.Context) {
	fmt.Println("********* DeleteProfilePhoto BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	fmt.Printf("Deleting profile photo for community with ID: %d\n", communityID)

	// Delete profile photo
	err = c.communityService.DeleteProfilePhoto(ctx, communityID)
	if err != nil {
		fmt.Printf("Error deleting profile photo: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community not found or has no profile photo")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to delete profile photo").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* DeleteProfilePhoto BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "Profile photo deleted successfully"}))
}

// JoinCommunity godoc
// @Summary Join a community
// @Description The authenticated user joins the specified community
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "User joined community successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 409 {object} dto.APIResponse{error=dto.ErrorDetail} "User is already a participant"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/participants [post]
func (c *CommunityController) JoinCommunity(ctx *gin.Context) {
	fmt.Println("********* JoinCommunity BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Get current user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User ID not found in context")))
		return
	}

	// Convert to int64
	userID, ok := userIDInterface.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid user ID format")))
		return
	}

	fmt.Printf("User %d joining community with ID: %d\n", userID, communityID)

	// Join community
	err = c.communityService.JoinCommunity(ctx, communityID, userID)
	if err != nil {
		fmt.Printf("Error joining community: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community or user not found")))
		case errors.Is(err, apperrors.ErrConflict) || strings.Contains(err.Error(), "already a participant"):
			ctx.JSON(http.StatusConflict, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeConflict, "User is already a participant in this community")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to join community").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* JoinCommunity BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "User joined community successfully"}))
}

// LeaveCommunity godoc
// @Summary Leave a community
// @Description The authenticated user leaves the specified community
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=dto.SuccessResponse} "User left community successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 409 {object} dto.APIResponse{error=dto.ErrorDetail} "Lead cannot leave community"
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/participants [delete]
func (c *CommunityController) LeaveCommunity(ctx *gin.Context) {
	fmt.Println("********* LeaveCommunity BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Get current user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeUnauthorized, "User ID not found in context")))
		return
	}

	// Convert to int64
	userID, ok := userIDInterface.(int64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Invalid user ID format")))
		return
	}

	fmt.Printf("User %d leaving community with ID: %d\n", userID, communityID)

	// Leave community
	err = c.communityService.LeaveCommunity(ctx, communityID, userID)
	if err != nil {
		fmt.Printf("Error leaving community: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community or user not found or user is not a participant")))
		case errors.Is(err, apperrors.ErrConflict) || strings.Contains(err.Error(), "lead cannot leave"):
			ctx.JSON(http.StatusConflict, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeConflict, "Lead cannot leave the community. Assign a new lead first.")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to leave community").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* LeaveCommunity BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(
		dto.SuccessResponse{Message: "User left community successfully"}))
}

// GetCommunityParticipants godoc
// @Summary Get community participants
// @Description Get all participants of a community
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Success 200 {object} dto.APIResponse{data=[]dto.CommunityParticipantResponse} "Participants retrieved successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/participants [get]
func (c *CommunityController) GetCommunityParticipants(ctx *gin.Context) {
	fmt.Println("********* GetCommunityParticipants BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	fmt.Printf("Getting participants for community with ID: %d\n", communityID)

	// Get participants
	participants, err := c.communityService.GetCommunityParticipants(ctx, communityID)
	if err != nil {
		fmt.Printf("Error getting community participants: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community not found")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to get community participants").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* GetCommunityParticipants BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(participants))
}

// IsUserParticipant godoc
// @Summary Check if user is a participant
// @Description Check if a specific user is a participant in the community
// @Tags communities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Community ID"
// @Param userId query int true "User ID"
// @Success 200 {object} dto.APIResponse{data=map[string]bool} "Participation status checked successfully"
// @Failure 400 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 401 {object} dto.APIResponse{error=dto.ErrorDetail} "Unauthorized: JWT token missing or invalid"
// @Failure 404 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Failure 500 {object} dto.APIResponse{error=dto.ErrorDetail}
// @Router /communities/{id}/participants/check [get]
func (c *CommunityController) IsUserParticipant(ctx *gin.Context) {
	fmt.Println("********* IsUserParticipant BAŞLANGIÇ *********")

	// Parse community ID from path
	idStr := ctx.Param("id")
	communityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid community ID")))
		return
	}

	// Parse user ID from query
	userIDStr := ctx.Query("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			dto.NewErrorDetail(dto.ErrorCodeInvalidRequest, "Invalid user ID")))
		return
	}

	fmt.Printf("Checking if user %d is a participant in community with ID: %d\n", userID, communityID)

	// Check if user is a participant
	isParticipant, err := c.communityService.IsUserParticipant(ctx, communityID, userID)
	if err != nil {
		fmt.Printf("Error checking participant status: %v\n", err)

		// Handle specific error cases
		switch {
		case errors.Is(err, apperrors.ErrResourceNotFound) || strings.Contains(err.Error(), "not found"):
			ctx.JSON(http.StatusNotFound, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeResourceNotFound, "Community or user not found")))
		default:
			ctx.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				dto.NewErrorDetail(dto.ErrorCodeInternalServer, "Failed to check participant status").WithDetails(err.Error())))
		}
		return
	}

	fmt.Println("********* IsUserParticipant BAŞARILI *********")
	ctx.JSON(http.StatusOK, dto.NewSuccessResponse(map[string]bool{
		"isParticipant": isParticipant,
	}))
}
