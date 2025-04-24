package dto

// UpdateTitleRequest represents a request to update an instructor's title
type UpdateTitleRequest struct {
	Title string `json:"title" binding:"required"`
}
