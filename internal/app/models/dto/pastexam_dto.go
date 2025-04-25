package dto

import "github.com/yigit/unisphere/internal/app/models"

// CreatePastExamRequest represents the request to create a past exam
type CreatePastExamRequest struct {
	Year         int     `json:"year" validate:"required,min=1900,max=2100" example:"2023"`                        // Year the exam was held (e.g., 2023)
	Term         string  `json:"term" validate:"required,oneof=FALL SPRING" example:"FALL"`                        // Term the exam was held (FALL or SPRING)
	DepartmentID int64   `json:"departmentId" validate:"required,min=1" example:"1"`                               // ID of the department for the course
	CourseCode   string  `json:"courseCode" validate:"required" example:"CENG301"`                                 // Course code (e.g., CENG301)
	Title        string  `json:"title" validate:"required" example:"Midterm Exam"`                                 // Title of the exam (e.g., Midterm 1, Final Exam)
	Content      string  `json:"content" validate:"required" example:"Exam content details..."`                    // Detailed content or description of the exam
	FileURL      *string `json:"fileUrl,omitempty" validate:"omitempty,url" example:"http://example.com/exam.pdf"` // Optional URL to the exam file (PDF, image, etc.)
}

// UpdatePastExamRequest represents the request to update a past exam
type UpdatePastExamRequest struct {
	Year         int     `json:"year" validate:"required,min=1900,max=2100" example:"2023"`                           // Year the exam was held
	Term         string  `json:"term" validate:"required,oneof=FALL SPRING" example:"FALL"`                           // Term the exam was held
	DepartmentID int64   `json:"departmentId" validate:"required,min=1" example:"1"`                                  // ID of the department
	CourseCode   string  `json:"courseCode" validate:"required" example:"CENG301"`                                    // Course code
	Title        string  `json:"title" validate:"required" example:"Midterm 1 - Updated"`                             // Title of the exam
	Content      string  `json:"content" validate:"required" example:"Updated exam content..."`                       // Detailed content
	FileURL      *string `json:"fileUrl,omitempty" validate:"omitempty,url" example:"http://example.com/exam_v2.pdf"` // Optional URL to the exam file
}

// PastExamResponse represents the response for a past exam including related entity names
type PastExamResponse struct {
	ID              int64  `json:"id" example:"10"`                                         // Unique identifier for the past exam
	Year            int    `json:"year" example:"2023"`                                     // Year the exam was held
	Term            string `json:"term" example:"FALL"`                                     // Term the exam was held (FALL or SPRING)
	FacultyID       int64  `json:"facultyId" example:"1"`                                   // ID of the faculty associated with the department
	FacultyName     string `json:"facultyName" example:"Engineering Faculty"`               // Name of the faculty
	DepartmentID    int64  `json:"departmentId" example:"1"`                                // ID of the department for the course
	DepartmentName  string `json:"departmentName" example:"Computer Engineering"`           // Name of the department
	CourseCode      string `json:"courseCode" example:"CENG301"`                            // Course code
	InstructorName  string `json:"instructorName" example:"Jane Smith"`                     // Name of the instructor who uploaded the exam
	Title           string `json:"title" example:"Midterm 1"`                               // Title of the exam
	Content         string `json:"content" example:"Exam content details..."`               // Detailed content or description of the exam
	FileURL         string `json:"fileUrl,omitempty" example:"http://example.com/exam.pdf"` // URL to the exam file, if available
	UploadedByEmail string `json:"uploadedByEmail" example:"instructor@school.edu.tr"`      // Email of the instructor who uploaded the exam
}

// PastExamListResponse represents the response for a list of past exams with pagination metadata
type PastExamListResponse struct {
	Exams      []PastExamResponse `json:"exams"`      // List of past exam details for the current page
	Pagination PaginationInfo     `json:"pagination"` // Pagination metadata
}

// PaginationInfo represents pagination metadata for list responses
/*
type PaginationInfo struct {
	CurrentPage int `json:"currentPage" example:"0"` // Current page number (0-based)
	TotalPages  int `json:"totalPages" example:"5"`  // Total number of pages available
	PageSize    int `json:"pageSize" example:"10"`   // Number of items per page
	TotalItems  int `json:"totalItems" example:"48"` // Total number of items matching the query
}
*/

// FromPastExam converts a model.PastExam to a PastExamResponse DTO
func FromPastExam(exam *models.PastExam) PastExamResponse {
	if exam == nil {
		return PastExamResponse{}
	}

	facultyName := ""
	if exam.Faculty != nil {
		facultyName = exam.Faculty.Name
	}

	departmentName := ""
	if exam.Department != nil {
		departmentName = exam.Department.Name
	}

	fileURL := ""
	if exam.FileURL != nil {
		fileURL = *exam.FileURL
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
		FileURL:         fileURL,
		UploadedByEmail: exam.UploadedByEmail,
	}
}
