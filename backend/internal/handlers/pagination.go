package handlers

import (
	"net/http"
	"strconv"
)

// Pagination holds parsed pagination parameters
type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

// DefaultPagination contains default values
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// ParsePagination extracts pagination parameters from request query string.
// Returns page (1-indexed), limit (capped at maxLimit), and calculated offset.
func ParsePagination(r *http.Request) Pagination {
	return ParsePaginationWithDefaults(r, DefaultLimit, MaxLimit)
}

// ParsePaginationWithDefaults extracts pagination with custom defaults.
func ParsePaginationWithDefaults(r *http.Request, defaultLimit, maxLimit int) Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = DefaultPage
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}

	offset := (page - 1) * limit

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
