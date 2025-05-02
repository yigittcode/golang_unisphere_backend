package dto

import (
	"time"
)

// --- Request DTOs ---

// CreateClassNoteRequest represents class note creation data
type CreateClassNoteRequest struct {
	CourseCode   string `json:"courseCode" form:"courseCode" binding:"required"`
	Title        string `json:"title" form:"title" binding:"required"`
	Description  string `json:"description" form:"description" binding:"required"`
	Content      string `json:"content" form:"content" binding:"required"`
	DepartmentID int64  `json:"departmentId" form:"departmentId" binding:"required,gt=0"`
}

// UpdateClassNoteRequest represents class note update data
type UpdateClassNoteRequest struct {
	CourseCode  string `json:"courseCode" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

// --- Response DTOs ---

// ClassNoteFileResponse represents complete file information for class notes
type ClassNoteFileResponse struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"fileName"`
	FileURL   string    `json:"fileUrl"`
	FileSize  int64     `json:"fileSize"`
	FileType  string    `json:"fileType"`
	CreatedAt time.Time `json:"createdAt"`
}

// SimpleClassNoteFileResponse represents just the file ID for class notes
// Used when returning file lists within note responses to minimize payload size
type SimpleClassNoteFileResponse struct {
	ID int64 `json:"id"`
}

// ClassNoteResponse represents basic class note information
type ClassNoteResponse struct {
	ID           int64                         `json:"id"`
	CourseCode   string                        `json:"courseCode"`
	Title        string                        `json:"title"`
	Description  string                        `json:"description"`
	Content      string                        `json:"content"`
	DepartmentID int64                         `json:"departmentId"`
	UserID       int64                         `json:"userId"`
	CreatedAt    time.Time                     `json:"createdAt"`
	UpdatedAt    time.Time                     `json:"updatedAt"`
	Files        []SimpleClassNoteFileResponse `json:"files,omitempty"`
}

// PaginationInfo is defined in response.go to avoid duplication

// ClassNoteListResponse represents a list of class notes
type ClassNoteListResponse struct {
	ClassNotes []ClassNoteResponse `json:"classNotes"`
	PaginationInfo
}

// ClassNoteFilterRequest represents class note filter parameters
type ClassNoteFilterRequest struct {
	DepartmentID *int64  `form:"departmentId,omitempty"`
	CourseCode   *string `form:"courseCode,omitempty"`
	InstructorID *int64  `form:"instructorId,omitempty"`
	Page         int     `form:"page,default=1" binding:"min=1"`
	PageSize     int     `form:"pageSize,default=10" binding:"min=1,max=100"`
	SortBy       string  `form:"sortBy,default=created_at" binding:"omitempty,oneof=created_at updated_at title course_code"`
	SortOrder    string  `form:"sortOrder,default=desc" binding:"omitempty,oneof=asc desc"`
}

// --- Helper Functions ---

// Helper functions (FromServiceClassNoteResponse, FromRepoPaginationInfo, MapServiceNotesToDTO) are removed
// as the mapping will be handled in the controller to avoid import cycles.
