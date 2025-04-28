package dto

// --- Request DTOs ---

// CreateClassNoteRequest represents class note creation data
type CreateClassNoteRequest struct {
	CourseCode   string `json:"courseCode" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description" binding:"required"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
}

// UpdateClassNoteRequest represents class note update data
type UpdateClassNoteRequest struct {
	CourseCode  string `json:"courseCode" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// --- Response DTOs ---

// ClassNoteResponse represents basic class note information
type ClassNoteResponse struct {
	ID           int64  `json:"id"`
	CourseCode   string `json:"courseCode"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	FileID       string `json:"fileId"`
	DepartmentID int64  `json:"departmentId"`
	InstructorID int64  `json:"instructorId"`
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
}

// --- Helper Functions ---

// Helper functions (FromServiceClassNoteResponse, FromRepoPaginationInfo, MapServiceNotesToDTO) are removed
// as the mapping will be handled in the controller to avoid import cycles.
