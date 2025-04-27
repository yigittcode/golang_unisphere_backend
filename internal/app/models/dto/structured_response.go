package dto

import "time"

// StructuredResponse provides a base structured API response with nested objects
type StructuredResponse struct {
	Success   bool         `json:"success" example:"true"`
	Message   string       `json:"message" example:"Operation completed successfully"`
	Data      interface{}  `json:"data,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Timestamp time.Time    `json:"timestamp" example:"2025-04-23T12:01:05.123Z"`
}

// DepartmentData represents department information in a structured format
type DepartmentData struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"Computer Engineering"`
	FacultyID int64  `json:"facultyId" example:"1"`
}

// FacultyData represents faculty information in a structured format
type FacultyData struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Engineering Faculty"`
}

// FileData represents file information in a structured format
type FileData struct {
	ID           int64  `json:"id" example:"123"`
	FileName     string `json:"fileName" example:"lecture_slides.pdf"`
	FileURL      string `json:"fileUrl" example:"http://example.com/uploads/123"`
	FileSize     int64  `json:"fileSize" example:"1048576"`
	FileType     string `json:"fileType" example:"application/pdf"`
	ResourceType string `json:"resourceType" example:"PAST_EXAM"`
	CreatedAt    string `json:"createdAt" example:"2024-01-15T10:00:00Z"`
}

// UserData represents user information in a structured format
type UserData struct {
	ID         int64           `json:"id" example:"1"`
	Email      string          `json:"email" example:"user@school.edu.tr"`
	FirstName  string          `json:"firstName" example:"John"`
	LastName   string          `json:"lastName" example:"Doe"`
	RoleType   string          `json:"roleType" example:"STUDENT" enums:"STUDENT,INSTRUCTOR"`
	Profile    ProfileData     `json:"profile,omitempty"`
	Department *DepartmentData `json:"department,omitempty"`
	Faculty    *FacultyData    `json:"faculty,omitempty"`
}

// ProfileData represents profile-specific information based on role
type ProfileData struct {
	ProfilePhoto *FileData `json:"profilePhoto,omitempty"`
	// Student specific fields
	Identifier     *string `json:"identifier,omitempty" example:"12345678"`
	GraduationYear *int    `json:"graduationYear,omitempty" example:"2025"`
	// Instructor specific fields
	Title *string `json:"title,omitempty" example:"Professor"`
}

// NewStructuredResponse creates a standard structured API response
func NewStructuredResponse(data interface{}, message string) StructuredResponse {
	return StructuredResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewStructuredUserResponse creates a response with properly structured user data
func NewStructuredUserResponse(profile *UserProfile) *UserData {
	if profile == nil {
		return nil
	}

	userData := &UserData{
		ID:        profile.ID,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		RoleType:  profile.RoleType,
		Profile:   ProfileData{},
	}

	// Add profile photo if exists
	if profile.ProfilePhotoFileId != nil {
		userData.Profile.ProfilePhoto = &FileData{
			ID: *profile.ProfilePhotoFileId,
		}
	}

	// Add student-specific fields
	if profile.Identifier != nil {
		userData.Profile.Identifier = profile.Identifier
	}
	if profile.GraduationYear != nil {
		userData.Profile.GraduationYear = profile.GraduationYear
	}

	// Add instructor-specific fields
	if profile.Title != nil {
		userData.Profile.Title = profile.Title
	}

	// Add department if exists
	if profile.DepartmentID > 0 {
		userData.Department = &DepartmentData{
			ID:   profile.DepartmentID,
			Name: profile.DepartmentName,
		}

		// Add faculty if exists
		if profile.FacultyID > 0 {
			userData.Faculty = &FacultyData{
				ID:   profile.FacultyID,
				Name: profile.FacultyName,
			}
		}
	}

	return userData
}

// PaginatedResponse represents a paginated list with metadata
type PaginatedResponse struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}
