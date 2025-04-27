package dto

// FileResponse represents the response for a file
type FileResponse struct {
	ID           int64  `json:"id" example:"123"`                                 // Unique identifier for the file
	FileName     string `json:"fileName" example:"lecture_slides.pdf"`            // Name of the file
	FileURL      string `json:"fileUrl" example:"http://example.com/uploads/123"` // URL to access the file
	FileSize     int64  `json:"fileSize" example:"1048576"`                       // Size of the file in bytes
	FileType     string `json:"fileType" example:"application/pdf"`               // MIME type of the file
	ResourceType string `json:"resourceType" example:"PAST_EXAM"`                 // Type of resource this file is attached to
	CreatedAt    string `json:"createdAt" example:"2024-01-15T10:00:00Z"`         // Timestamp when the file was created
}

// FilesResponse represents a collection of files
type FilesResponse struct {
	Files []FileResponse `json:"files"` // Collection of file responses
}
