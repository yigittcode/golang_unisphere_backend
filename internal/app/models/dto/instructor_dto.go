package dto

// RegisterInstructorRequest represents instructor registration data
type RegisterInstructorRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FirstName    string `json:"firstName" binding:"required"`
	LastName     string `json:"lastName" binding:"required"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
	Title        string `json:"title" binding:"required"`
}

// InstructorResponse extends UserResponse with instructor-specific fields
type InstructorResponse struct {
	UserResponse
	Title string `json:"title"`
}

// UpdateTitleRequest represents the request to update an instructor's title
type UpdateTitleRequest struct {
	Title string `json:"title" binding:"required" example:"Professor"`
}
