package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success sends a successful response with data
func Success(ctx *gin.Context, httpCode int, data interface{}) {
	ctx.JSON(httpCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessWithMeta sends a successful response with data and metadata
func SuccessWithMeta(ctx *gin.Context, httpCode int, data interface{}, meta interface{}) {
	ctx.JSON(httpCode, gin.H{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}

// Common success responses for convenience
func OK(ctx *gin.Context, data interface{}) {
	Success(ctx, http.StatusOK, data)
}

func Created(ctx *gin.Context, data interface{}) {
	Success(ctx, http.StatusCreated, data)
}

func OKWithMeta(ctx *gin.Context, data interface{}, meta interface{}) {
	SuccessWithMeta(ctx, http.StatusOK, data, meta)
}

// Pagination metadata structure (optional helper)
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(page, limit int, total int64) PaginationMeta {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
