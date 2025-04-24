package dto

import "github.com/yigit/unisphere/internal/app/models"

// CreatePastExamRequest represents the request to create a past exam
type CreatePastExamRequest struct {
	Year         int    `json:"year" validate:"required,min=1900,max=2100"`
	Term         string `json:"term" validate:"required,oneof=FALL SPRING"`
	DepartmentID int64  `json:"departmentId" validate:"required,min=1"`
	CourseCode   string `json:"courseCode" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content" validate:"required"`
	FileURL      string `json:"fileUrl" validate:"omitempty,url"`
}

// UpdatePastExamRequest represents the request to update a past exam
type UpdatePastExamRequest struct {
	Year         int    `json:"year" validate:"required,min=1900,max=2100"`
	Term         string `json:"term" validate:"required,oneof=FALL SPRING"`
	DepartmentID int64  `json:"departmentId" validate:"required,min=1"`
	CourseCode   string `json:"courseCode" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content" validate:"required"`
	FileURL      string `json:"fileUrl" validate:"omitempty,url"`
}

// PastExamResponse represents the response for a past exam
type PastExamResponse struct {
	ID              int64  `json:"id"`
	Year            int    `json:"year"`
	Term            string `json:"term"`
	FacultyID       int64  `json:"facultyId"`
	FacultyName     string `json:"facultyName"`
	DepartmentID    int64  `json:"departmentId"`
	DepartmentName  string `json:"departmentName"`
	CourseCode      string `json:"courseCode"`
	InstructorName  string `json:"instructorName"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	FileURL         string `json:"photo"` // Using "photo" as specified in the API spec
	UploadedByEmail string `json:"uploadedByEmail"`
}

// PastExamListResponse represents the response for a list of past exams with pagination
type PastExamListResponse struct {
	Exams      []PastExamResponse `json:"exams"`
	Pagination PaginationInfo     `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
	PageSize    int `json:"pageSize"`
	TotalItems  int `json:"totalItems"`
}

// FromPastExam converts a model.PastExam to a PastExamResponse
func FromPastExam(exam *models.PastExam) PastExamResponse {
	if exam == nil {
		return PastExamResponse{}
	}

	facultyName := ""
	departmentName := ""

	if exam.Faculty != nil {
		facultyName = exam.Faculty.Name
	}

	if exam.Department != nil {
		departmentName = exam.Department.Name
	}

	return PastExamResponse{
		ID:              exam.ID,
		Year:            exam.Year,
		Term:            string(exam.Term),
		FacultyID:       exam.FacultyID,
		FacultyName:     facultyName,
		DepartmentID:    exam.DepartmentID,
		DepartmentName:  departmentName,
		CourseCode:      exam.CourseCode,
		InstructorName:  exam.UploadedByName,
		Title:           exam.Title,
		Content:         exam.Content,
		FileURL:         exam.FileURL,
		UploadedByEmail: exam.UploadedByEmail,
	}
}
