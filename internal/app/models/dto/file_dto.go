package dto

// FileType represents the type of file
type FileType string

// Standard file types in the application
const (
	FileTypePastExam     FileType = "PAST_EXAM"
	FileTypeClassNote    FileType = "CLASS_NOTE"
	FileTypeProfilePhoto FileType = "PROFILE_PHOTO"
)

// FileResponse represents basic file information
// @Description Basic file information
type FileResponse struct {
	ID   string `json:"id" example:"123456"`
	Name string `json:"name" example:"document.pdf"`
	Type string `json:"type" example:"application/pdf"`
}

// FileUploadResponse represents file upload result
// @Description Result information after successful file upload
type FileUploadResponse struct {
	FileID string `json:"fileId" example:"123456"`
}
