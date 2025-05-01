package dto

import "time"

// Term represents a semester term for past exams
type Term string

// Term constants
const (
	TermFall   Term = "FALL"
	TermSpring Term = "SPRING"
)

// PastExamFileResponse represents file information specific to past exams
type PastExamFileResponse struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"fileName"`
	FileURL   string    `json:"fileUrl"`
	FileSize  int64     `json:"fileSize"`
	FileType  string    `json:"fileType"`
	CreatedAt time.Time `json:"createdAt"`
}

// PastExamResponse represents basic past exam information
type PastExamResponse struct {
	ID           int64                  `json:"id"`
	CourseCode   string                 `json:"courseCode"`
	Year         int                    `json:"year"`
	Term         string                 `json:"term"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	DepartmentID int64                  `json:"departmentId"`
	InstructorID int64                  `json:"instructorId"`
	Files        []PastExamFileResponse `json:"files,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// CreatePastExamRequest represents past exam creation data
type CreatePastExamRequest struct {
	CourseCode   string `json:"courseCode" binding:"required"`
	Year         int    `json:"year" binding:"required,gt=1900"`
	Term         string `json:"term" binding:"required,oneof=FALL SPRING"`
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
}

// UpdatePastExamRequest represents past exam update data
type UpdatePastExamRequest struct {
	CourseCode string `json:"courseCode" binding:"required"`
	Year       int    `json:"year" binding:"required,gt=1900"`
	Term       string `json:"term" binding:"required,oneof=FALL SPRING"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// PastExamListResponse represents a list of past exams
type PastExamListResponse struct {
	PastExams []PastExamResponse `json:"pastExams"`
	PaginationInfo
}

// PastExamFilterRequest represents past exam filter parameters
type PastExamFilterRequest struct {
	DepartmentID *int64  `form:"departmentId,omitempty"`
	CourseCode   *string `form:"courseCode,omitempty"`
	Year         *int    `form:"year,omitempty"`
	Term         *string `form:"term,omitempty"`
	InstructorID *int64  `form:"instructorId,omitempty"`
	Page         int     `form:"page,default=1" binding:"min=1"`
	PageSize     int     `form:"pageSize,default=10" binding:"min=1,max=100"`
}
