package helpers

import (
	"math"

	"github.com/yigit/unisphere/internal/app/models/dto" // Import DTO for PaginationInfo
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	DefaultPage     = 0 // Default page is 0-based
)

// CalculateOffsetLimit calculates the offset and limit for SQL queries based on 0-based page index.
func CalculateOffsetLimit(page, size int) (offset uint64, limit int) {
	if size <= 0 || size > MaxPageSize {
		limit = DefaultPageSize
	} else {
		limit = size
	}

	if page < 0 {
		page = DefaultPage
	}

	offset = uint64(page * limit)
	return offset, limit
}

// NewPaginationInfo creates a standard PaginationInfo DTO.
// page should be the 0-based page number used in the query.
func NewPaginationInfo(totalItems int64, page, size int) dto.PaginationInfo {
	if size <= 0 {
		size = DefaultPageSize
	}
	if page < 0 {
		page = DefaultPage
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(size)))
	}

	return dto.PaginationInfo{
		CurrentPage: page + 1, // Return 1-based page for the API response
		TotalPages:  totalPages,
		PageSize:    size,
		TotalItems:  totalItems,
	}
}
