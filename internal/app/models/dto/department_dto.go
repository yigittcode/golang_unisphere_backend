package dto

// DepartmentResponse represents basic department information
type DepartmentResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	FacultyID int64  `json:"facultyId"`
}

// CreateDepartmentRequest represents department creation data
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	FacultyID int64  `json:"facultyId" binding:"required,gt=0"`
}

// UpdateDepartmentRequest represents department update data
type UpdateDepartmentRequest struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

// DepartmentListResponse represents a list of departments
type DepartmentListResponse struct {
	Departments []DepartmentResponse `json:"departments"`
	PaginationInfo
}
