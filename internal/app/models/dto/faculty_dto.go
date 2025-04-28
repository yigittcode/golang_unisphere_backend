package dto

// FacultyResponse represents basic faculty information
type FacultyResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateFacultyRequest represents faculty creation data
type CreateFacultyRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateFacultyRequest represents faculty update data
type UpdateFacultyRequest struct {
	Name string `json:"name" binding:"required"`
}

// FacultyListResponse represents a list of faculties
type FacultyListResponse struct {
	Faculties []FacultyResponse `json:"faculties"`
	PaginationInfo
}
 