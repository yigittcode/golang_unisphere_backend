package dto

// RegisterStudentRequest represents student registration data
type RegisterStudentRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	DepartmentID   int64  `json:"departmentId" binding:"required,gt=0"`
	StudentID      string `json:"studentId" binding:"required,len=8,numeric"`
	GraduationYear *int   `json:"graduationYear,omitempty" binding:"omitempty,min=1900"`
}

// StudentResponse extends UserResponse with student-specific fields
type StudentResponse struct {
	UserResponse
	StudentID      string `json:"studentId"`
	GraduationYear *int   `json:"graduationYear,omitempty"`
}
