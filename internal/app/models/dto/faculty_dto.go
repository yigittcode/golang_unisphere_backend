package dto

// FacultyResponse represents basic faculty information
type FacultyResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// CreateFacultyRequest represents faculty creation data
type CreateFacultyRequest struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

// UpdateFacultyRequest represents faculty update data
type UpdateFacultyRequest struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

// FacultyListResponse represents a list of faculties
type FacultyListResponse struct {
	Faculties []FacultyResponse `json:"faculties"`
	PaginationInfo
}
