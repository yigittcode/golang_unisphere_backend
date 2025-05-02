package models

import "time"

// Community represents a student or academic community
type Community struct {
	ID                int64     `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Abbreviation      string    `json:"abbreviation" db:"abbreviation"`
	LeadID            int64     `json:"leadId" db:"lead_id"`
	ProfilePhotoFileID *int64    `json:"profilePhotoFileId,omitempty" db:"profile_photo_file_id"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
	
	// Related entities
	Lead         *User          `json:"lead,omitempty"`
	ProfilePhoto *File          `json:"profilePhoto,omitempty"`
	Participants []*User        `json:"participants,omitempty"`
	Files        []*File        `json:"files,omitempty"`
}

// CommunityParticipant represents a user participating in a community
type CommunityParticipant struct {
	ID          int64     `json:"id" db:"id"`
	CommunityID int64     `json:"communityId" db:"community_id"`
	UserID      int64     `json:"userId" db:"user_id"`
	JoinedAt    time.Time `json:"joinedAt" db:"joined_at"`
	
	// Related entities
	User *User `json:"user,omitempty"`
}