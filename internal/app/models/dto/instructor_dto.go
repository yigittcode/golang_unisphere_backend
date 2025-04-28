package dto

// InstructorResponse represents basic instructor information
type InstructorResponse struct {
	ID             int64  `json:"id"`                       // Unique identifier for the instructor
	UserID         int64  `json:"userId"`                   // Associated user ID
	FirstName      string `json:"firstName"`                // Instructor's first name
	LastName       string `json:"lastName"`                 // Instructor's last name
	Email          string `json:"email"`                    // Instructor's email
	Title          string `json:"title"`                    // Instructor's academic title
	DepartmentID   int64  `json:"departmentId"`             // Department ID
	DepartmentName string `json:"departmentName,omitempty"` // Department name
	FacultyName    string `json:"facultyName,omitempty"`    // Faculty name
	CreatedAt      string `json:"createdAt,omitempty"`      // When instructor account was created
}

// InstructorListResponse represents a list of instructors
type InstructorListResponse struct {
	Instructors []InstructorResponse `json:"instructors"`
	PaginationInfo
}

// InstructorFilterRequest represents instructor filter parameters
type InstructorFilterRequest struct {
	DepartmentID *int64  `form:"departmentId,omitempty"`
	Name         *string `form:"name,omitempty"`
	Title        *string `form:"title,omitempty"`
	Page         int     `form:"page,default=1" binding:"min=1"`
	PageSize     int     `form:"pageSize,default=10" binding:"min=1,max=100"`
}

// InstructorsResponse represents a list of instructors
type InstructorsResponse struct {
	Instructors []InstructorResponse `json:"instructors"` // List of instructors
}

// UpdateTitleRequest represents the request to update an instructor's title
type UpdateTitleRequest struct {
	Title string `json:"title" validate:"required" example:"Professor"` // New academic title
}
