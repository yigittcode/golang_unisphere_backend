package dto

// --- Request DTOs ---

// CreateClassNoteRequest represents the data needed to create a new class note.
// Based on services.CreateClassNoteRequest
type CreateClassNoteRequest struct {
	Year         int    `json:"year" validate:"required,gte=2000" example:"2024"`                                 // Year the note corresponds to
	Term         string `json:"term" validate:"required,oneof=FALL SPRING" example:"SPRING"`                      // Term (FALL or SPRING)
	DepartmentID int64  `json:"departmentId" validate:"required,gt=0" example:"1"`                                // ID of the department for the course
	CourseCode   string `json:"courseCode" validate:"required,alphanum,uppercase,min=3,max=10" example:"CENG304"` // Course code (e.g., CENG304)
	Title        string `json:"title" validate:"required,min=5,max=255" example:"Lecture Notes - Week 5"`         // Title of the class note
	Content      string `json:"content" validate:"required,min=10" example:"Detailed notes covering topic X..."`  // Main content of the note
	// Image field removed, file should be sent via multipart/form-data
}

// UpdateClassNoteRequest represents the data needed to update a class note.
// Based on services.UpdateClassNoteRequest
type UpdateClassNoteRequest struct {
	Year         int    `json:"year" validate:"required,gte=2000" example:"2024"`                                        // Year the note corresponds to
	Term         string `json:"term" validate:"required,oneof=FALL SPRING" example:"SPRING"`                             // Term (FALL or SPRING)
	DepartmentID int64  `json:"departmentId" validate:"required,gt=0" example:"1"`                                       // ID of the department for the course
	CourseCode   string `json:"courseCode" validate:"required,alphanum,uppercase,min=3,max=10" example:"CENG304"`        // Course code (e.g., CENG304)
	Title        string `json:"title" validate:"required,min=5,max=255" example:"Updated Lecture Notes - Week 5"`        // Title of the class note
	Content      string `json:"content" validate:"required,min=10" example:"Updated detailed notes covering topic X..."` // Main content of the note
	// Image field removed, file should be sent via multipart/form-data
}

// --- Response DTOs ---

// ClassNoteResponse represents the data returned for a single class note.
// Based on services.ClassNoteResponse
type ClassNoteResponse struct {
	ID                int64          `json:"id" example:"15"`                                      // Unique identifier for the class note
	Year              int            `json:"year" example:"2024"`                                  // Year the note corresponds to
	Term              string         `json:"term" example:"SPRING"`                                // Term (FALL or SPRING)
	FacultyID         int64          `json:"facultyId" example:"1"`                                // ID of the faculty associated with the department
	FacultyName       string         `json:"facultyName" example:"Engineering Faculty"`            // Name of the faculty
	DepartmentID      int64          `json:"departmentId" example:"1"`                             // ID of the department for the course
	DepartmentName    string         `json:"departmentName" example:"Computer Engineering"`        // Name of the department
	CourseCode        string         `json:"courseCode" example:"CENG304"`                         // Course code
	Title             string         `json:"title" example:"Lecture Notes - Week 5"`               // Title of the class note
	Content           string         `json:"content" example:"Detailed notes covering topic X..."` // Main content of the note
	Files             []FileResponse `json:"files,omitempty"`                                      // Files attached to the note (new field for multiple files)
	UploaderName      string         `json:"uploaderName" example:"John Doe"`                      // Name of the user who uploaded the note
	UploaderEmail     string         `json:"uploaderEmail" example:"john.doe@example.com"`         // Email of the user who uploaded the note
	UploadedByStudent bool           `json:"uploadedByStudent" example:"true"`                     // True if uploaded by a student, false if by an instructor
	CreatedAt         string         `json:"createdAt" example:"2024-01-15T10:00:00Z"`             // Timestamp when the note was created
	UpdatedAt         string         `json:"updatedAt" example:"2024-01-16T11:30:00Z"`             // Timestamp when the note was last updated
}

// PaginationInfo is defined in response.go to avoid duplication

// ClassNoteListResponse represents the response for a list of class notes with pagination metadata.
type ClassNoteListResponse struct {
	Notes      []ClassNoteResponse `json:"notes"`      // List of class note details for the current page
	Pagination PaginationInfo      `json:"pagination"` // Pagination metadata
}

// --- Helper Functions ---

// Helper functions (FromServiceClassNoteResponse, FromRepoPaginationInfo, MapServiceNotesToDTO) are removed
// as the mapping will be handled in the controller to avoid import cycles.
