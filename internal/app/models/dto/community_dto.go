package dto

import "time"

// --- Request DTOs ---

// CreateCommunityRequest represents community creation data
type CreateCommunityRequest struct {
	Name         string `json:"name" form:"name" binding:"required"`
	Abbreviation string `json:"abbreviation" form:"abbreviation" binding:"required"`
	// ProfilePhoto is handled separately in the multipart form
}

// UpdateCommunityRequest represents community update data
type UpdateCommunityRequest struct {
	Name         string `json:"name" form:"name" binding:"required"`
	Abbreviation string `json:"abbreviation" form:"abbreviation" binding:"required"`
	LeadID       int64  `json:"leadId" form:"leadId" binding:"required,gt=0"`
}

// JoinCommunityRequest represents the request to join a community
type JoinCommunityRequest struct {
	// Empty struct, uses authenticated user
}

// LeaveCommunityRequest represents the request to leave a community
type LeaveCommunityRequest struct {
	// Empty struct, uses authenticated user
}

// --- Response DTOs ---

// ProfilePhotoResponse represents a profile photo for communities
type ProfilePhotoResponse struct {
	ID      int64  `json:"id"`
	FileURL string `json:"fileUrl"`
}

// CommunityParticipantResponse represents a participant in a community
type CommunityParticipantResponse struct {
	UserID   int64     `json:"userId"`
	JoinedAt time.Time `json:"joinedAt"`
}

// CommunityResponse represents basic community information
type CommunityResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Abbreviation       string  `json:"abbreviation"`
	LeadID             int64   `json:"leadId"`
	ProfilePhotoFileID *int64  `json:"profilePhotoFileId,omitempty"`
	ProfilePhotoURL    *string `json:"profilePhotoUrl,omitempty"`
	ParticipantCount   int     `json:"participantCount,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// CommunityDetailResponse extends CommunityResponse with participant details
type CommunityDetailResponse struct {
	CommunityResponse
	Participants []CommunityParticipantResponse `json:"participants,omitempty"`
}

// Note: SimpleCommunityFileResponse has been removed as file management is now handled through chat

// CommunityListResponse represents a list of communities
type CommunityListResponse struct {
	Communities []CommunityResponse `json:"communities"`
	PaginationInfo
}

// CommunityFilterRequest represents community filter parameters
type CommunityFilterRequest struct {
	LeadID   *int64  `form:"leadId,omitempty"`
	Search   *string `form:"search,omitempty"` // For searching by name or abbreviation
	Page     int     `form:"page,default=1" binding:"min=1"`
	PageSize int     `form:"pageSize,default=10" binding:"min=1,max=100"`
}

// UserBasicResponse represents minimal user information for including in community responses
type UserBasicResponse struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}
