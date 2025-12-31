// Package pagination provides cursor-based and offset-based pagination utilities.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Cursor represents a pagination cursor containing the ID and timestamp of the last item.
type Cursor struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// Params holds pagination parameters from a request.
type Params struct {
	// Cursor-based pagination
	Cursor    string
	Direction string // "next" or "prev"

	// Offset-based pagination (fallback)
	Page  int
	Limit int

	// Parsed cursor data
	ParsedCursor *Cursor
}

// Result represents a paginated response.
type Result struct {
	// Data is set by the caller
	Data interface{} `json:"data"`

	// Cursor-based pagination
	NextCursor string `json:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`

	// Offset-based pagination (for backwards compatibility)
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"totalPages,omitempty"`
}

// DefaultLimit is the default number of items per page.
const DefaultLimit = 20

// MaxLimit is the maximum number of items per page.
const MaxLimit = 100

// ParseParams extracts pagination parameters from an HTTP request.
// It supports both cursor-based and offset-based pagination.
func ParseParams(r *http.Request) (*Params, error) {
	q := r.URL.Query()

	params := &Params{
		Cursor:    q.Get("cursor"),
		Direction: q.Get("direction"),
		Limit:     DefaultLimit,
		Page:      1,
	}

	// Parse limit
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return nil, errors.New("invalid limit parameter")
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}
		params.Limit = limit
	}

	// If cursor is provided, decode it
	if params.Cursor != "" {
		cursor, err := DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		params.ParsedCursor = cursor
	}

	// Parse page for offset-based pagination (fallback)
	if pageStr := q.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page parameter")
		}
		params.Page = page
	}

	// Default direction is "next"
	if params.Direction == "" {
		params.Direction = "next"
	}

	return params, nil
}

// EncodeCursor encodes a cursor to a base64 string.
func EncodeCursor(id uint, createdAt time.Time) string {
	cursor := Cursor{
		ID:        id,
		CreatedAt: createdAt,
	}
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a base64 cursor string to a Cursor struct.
func DecodeCursor(encoded string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	return &cursor, nil
}

// IsCursorBased returns true if the request uses cursor-based pagination.
func (p *Params) IsCursorBased() bool {
	return p.Cursor != "" || p.ParsedCursor != nil
}

// Offset returns the offset for offset-based pagination.
func (p *Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// NewResult creates a new pagination result.
func NewResult(data interface{}, total int, params *Params) *Result {
	result := &Result{
		Data:  data,
		Total: total,
		Limit: params.Limit,
	}

	if !params.IsCursorBased() {
		// Offset-based pagination
		result.Page = params.Page
		result.TotalPages = (total + params.Limit - 1) / params.Limit
		result.HasMore = params.Page < result.TotalPages
	}

	return result
}

// SetCursors sets the next and previous cursors based on the result set.
// items should be a slice of structs with ID and CreatedAt fields.
func (r *Result) SetCursors(hasMore bool, lastID uint, lastCreatedAt time.Time, firstID uint, firstCreatedAt time.Time) {
	r.HasMore = hasMore

	if hasMore && lastID > 0 {
		r.NextCursor = EncodeCursor(lastID, lastCreatedAt)
	}

	if firstID > 0 {
		r.PrevCursor = EncodeCursor(firstID, firstCreatedAt)
	}
}

// CursorQuery represents the SQL query parts for cursor-based pagination.
type CursorQuery struct {
	Where string
	Args  []interface{}
	Order string
}

// BuildCursorQuery builds the WHERE clause and ORDER BY for cursor-based pagination.
// This uses a compound cursor on (created_at, id) for stable ordering.
func BuildCursorQuery(cursor *Cursor, direction string) *CursorQuery {
	query := &CursorQuery{}

	if cursor == nil {
		// No cursor - start from beginning
		query.Order = "created_at DESC, id DESC"
		return query
	}

	if direction == "prev" {
		// Going backwards - get items newer than cursor
		query.Where = "(created_at > ? OR (created_at = ? AND id > ?))"
		query.Args = []interface{}{cursor.CreatedAt, cursor.CreatedAt, cursor.ID}
		query.Order = "created_at ASC, id ASC"
	} else {
		// Going forwards (default) - get items older than cursor
		query.Where = "(created_at < ? OR (created_at = ? AND id < ?))"
		query.Args = []interface{}{cursor.CreatedAt, cursor.CreatedAt, cursor.ID}
		query.Order = "created_at DESC, id DESC"
	}

	return query
}
