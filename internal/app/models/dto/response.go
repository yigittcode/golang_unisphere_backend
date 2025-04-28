package dto

import "time"

// APIResponse is the generic structure for all API responses.
// @Description Generic response structure for all API endpoints
type APIResponse struct {
	// Response payload - specific to each endpoint
	Data interface{} `json:"data,omitempty"`

	// Error details - only present when Success is false
	Error *ErrorDetail `json:"error,omitempty" swaggerignore:"true"`

	// Timestamp of when the response was generated
	Timestamp time.Time `json:"timestamp" example:"2025-04-23T12:01:05.123Z"`
}

// SuccessResponse represents a standard success response message for API endpoints
// Often used as Data in APIResponse when no other specific data is needed (e.g., for Delete operations)
type SuccessResponse struct {
	Message string `json:"message" example:"Resource deleted successfully"`
}

// NewSuccessResponse creates a standard success API response.
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Data:      data,
		Timestamp: time.Now(),
	}
}

// PaginationInfo represents pagination metadata for list responses
type PaginationInfo struct {
	CurrentPage int   `json:"currentPage" example:"0"` // Current page number (0-based)
	TotalPages  int   `json:"totalPages" example:"5"`  // Total number of pages available
	PageSize    int   `json:"pageSize" example:"10"`   // Number of items per page
	TotalItems  int64 `json:"totalItems" example:"48"` // Total number of items matching the query
}
