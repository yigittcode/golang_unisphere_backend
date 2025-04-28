package dto

// InstructorResponse represents the response for an instructor
type InstructorResponse struct {
	ID            int64  `json:"id" example:"1"`                          // Unique identifier for the instructor
	UserID        int64  `json:"userId" example:"5"`                      // Associated user ID
	FirstName     string `json:"firstName" example:"John"`                // Instructor's first name
	LastName      string `json:"lastName" example:"Doe"`                  // Instructor's last name
	Email         string `json:"email" example:"john.doe@school.edu.tr"`  // Instructor's email
	Title         string `json:"title" example:"Associate Professor"`     // Instructor's academic title
	DepartmentID  int64  `json:"departmentId" example:"2"`                // Department ID
	DepartmentName string `json:"departmentName" example:"Computer Engineering"` // Department name
	FacultyName   string `json:"facultyName,omitempty" example:"Engineering Faculty"` // Faculty name
	CreatedAt     string `json:"createdAt" example:"2024-01-15T10:00:00Z"` // When instructor account was created
}

// InstructorsResponse represents a list of instructors
type InstructorsResponse struct {
	Instructors []InstructorResponse `json:"instructors"` // List of instructors
}

// UpdateTitleRequest represents the request to update an instructor's title
type UpdateTitleRequest struct {
	Title string `json:"title" validate:"required" example:"Professor"` // New academic title
}