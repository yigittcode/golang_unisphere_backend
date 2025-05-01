package helpers

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/models/dto" // Import DTO for PaginationInfo
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	DefaultPage     = 1 // Default page is 1-based
)

// CalculateOffsetLimit calculates the offset and limit for SQL queries based on 1-based page index.
func CalculateOffsetLimit(page, size int) (offset uint64, limit int) {
	if size <= 0 || size > MaxPageSize {
		limit = DefaultPageSize
	} else {
		limit = size
	}

	if page < 1 {
		page = DefaultPage
	}

	// 1-tabanlı sayfa numarasını 0-tabanlı offset'e çevir
	offset = uint64((page - 1) * limit)
	return offset, limit
}

// NewPaginationInfo creates a standard PaginationInfo DTO.
// page should be the 1-based page number.
func NewPaginationInfo(totalItems int64, page, size int) dto.PaginationInfo {
	if size <= 0 {
		size = DefaultPageSize
	}
	if page < 1 {
		page = DefaultPage
	}

	// Calculate total pages based on total items
	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(size)))
	} else {
		// If no items, set totalPages to 1 when we're on page 1, otherwise keep it at 0
		if page == 1 {
			totalPages = 1
		}
	}

	// Ensure currentPage never exceeds totalPages
	currentPage := page
	if totalPages > 0 && currentPage > totalPages {
		currentPage = totalPages
	}

	return dto.PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		PageSize:    size,
		TotalItems:  totalItems,
	}
}

// ParsePaginationParams extracts and validates pagination parameters from the request
func ParsePaginationParams(c *gin.Context) (page, size int) {
	// Extract page parameter (API uses 1-based)
	pageStr := c.DefaultQuery("page", "1") // Default is page 1
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = DefaultPage // Default to page 1 if invalid
	}

	// Extract size parameter
	sizeStr := c.DefaultQuery("size", "10")
	size, err = strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > MaxPageSize {
		size = DefaultPageSize
	}

	return page, size
}

// CalculateSliceIndices calculates the start and end indices for slicing an array for pagination
func CalculateSliceIndices(page, size, totalItems int) (start, end int) {
	if size <= 0 {
		size = DefaultPageSize
	}
	if page < 1 {
		page = DefaultPage
	}

	// 1-tabanlı sayfa numarasını 0-tabanlı dizin indeksine çevir
	start = (page - 1) * size
	end = start + size

	if start >= totalItems {
		start = totalItems
		end = totalItems
	}

	return start, end
}
