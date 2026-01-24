package pagination

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

// ============ Constants Tests ============

func TestDefaultLimit(t *testing.T) {
	if DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d, want 20", DefaultLimit)
	}
}

func TestMaxLimit(t *testing.T) {
	if MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want 100", MaxLimit)
	}
}

// ============ Cursor Encoding/Decoding Tests ============

func TestEncodeCursor(t *testing.T) {
	id := uint(42)
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	encoded := EncodeCursor(id, createdAt)

	if encoded == "" {
		t.Error("EncodeCursor() returned empty string")
	}

	// Verify it can be decoded
	cursor, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor() error = %v", err)
	}

	if cursor.ID != id {
		t.Errorf("Decoded cursor ID = %d, want %d", cursor.ID, id)
	}

	if !cursor.CreatedAt.Equal(createdAt) {
		t.Errorf("Decoded cursor CreatedAt = %v, want %v", cursor.CreatedAt, createdAt)
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := DecodeCursor("invalid-base64!!!")
	if err == nil {
		t.Error("DecodeCursor() should return error for invalid base64")
	}
}

func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	_, err := DecodeCursor("bm90LWpzb24=") // "not-json" in base64
	if err == nil {
		t.Error("DecodeCursor() should return error for invalid JSON")
	}
}

func TestEncodeCursor_Roundtrip(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		createdAt time.Time
	}{
		{"simple case", 1, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"large ID", 999999, time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"zero ID", 0, time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeCursor(tt.id, tt.createdAt)
			decoded, err := DecodeCursor(encoded)
			if err != nil {
				t.Fatalf("DecodeCursor() error = %v", err)
			}
			if decoded.ID != tt.id {
				t.Errorf("ID = %d, want %d", decoded.ID, tt.id)
			}
			if !decoded.CreatedAt.Equal(tt.createdAt) {
				t.Errorf("CreatedAt = %v, want %v", decoded.CreatedAt, tt.createdAt)
			}
		})
	}
}

// ============ ParseParams Tests ============

func TestParseParams_Defaults(t *testing.T) {
	req := createRequest(url.Values{})

	params, err := ParseParams(req)
	if err != nil {
		t.Fatalf("ParseParams() error = %v", err)
	}

	if params.Limit != DefaultLimit {
		t.Errorf("Limit = %d, want %d", params.Limit, DefaultLimit)
	}

	if params.Page != 1 {
		t.Errorf("Page = %d, want 1", params.Page)
	}

	if params.Direction != "next" {
		t.Errorf("Direction = %q, want 'next'", params.Direction)
	}
}

func TestParseParams_WithLimit(t *testing.T) {
	tests := []struct {
		name      string
		limitStr  string
		wantLimit int
		wantErr   bool
	}{
		{"valid limit", "50", 50, false},
		{"limit at max", "100", 100, false},
		{"limit over max", "200", 100, false}, // Should be capped to MaxLimit
		{"limit below min", "0", 0, true},
		{"negative limit", "-1", 0, true},
		{"invalid limit", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequest(url.Values{"limit": {tt.limitStr}})
			params, err := ParseParams(req)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseParams() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseParams() error = %v", err)
			}

			if params.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", params.Limit, tt.wantLimit)
			}
		})
	}
}

func TestParseParams_WithPage(t *testing.T) {
	tests := []struct {
		name     string
		pageStr  string
		wantPage int
		wantErr  bool
	}{
		{"valid page", "5", 5, false},
		{"page 1", "1", 1, false},
		{"page 0", "0", 0, true},
		{"negative page", "-1", 0, true},
		{"invalid page", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequest(url.Values{"page": {tt.pageStr}})
			params, err := ParseParams(req)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseParams() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseParams() error = %v", err)
			}

			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
		})
	}
}

func TestParseParams_WithCursor(t *testing.T) {
	validCursor := EncodeCursor(42, time.Now())

	req := createRequest(url.Values{"cursor": {validCursor}})
	params, err := ParseParams(req)

	if err != nil {
		t.Fatalf("ParseParams() error = %v", err)
	}

	if params.Cursor != validCursor {
		t.Errorf("Cursor = %q, want %q", params.Cursor, validCursor)
	}

	if params.ParsedCursor == nil {
		t.Error("ParsedCursor should not be nil")
	}

	if params.ParsedCursor.ID != 42 {
		t.Errorf("ParsedCursor.ID = %d, want 42", params.ParsedCursor.ID)
	}
}

func TestParseParams_WithInvalidCursor(t *testing.T) {
	req := createRequest(url.Values{"cursor": {"invalid-cursor"}})
	_, err := ParseParams(req)

	if err == nil {
		t.Error("ParseParams() should return error for invalid cursor")
	}
}

func TestParseParams_WithDirection(t *testing.T) {
	tests := []struct {
		name          string
		direction     string
		wantDirection string
	}{
		{"next", "next", "next"},
		{"prev", "prev", "prev"},
		{"empty defaults to next", "", "next"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			if tt.direction != "" {
				values["direction"] = []string{tt.direction}
			}
			req := createRequest(values)
			params, err := ParseParams(req)

			if err != nil {
				t.Fatalf("ParseParams() error = %v", err)
			}

			if params.Direction != tt.wantDirection {
				t.Errorf("Direction = %q, want %q", params.Direction, tt.wantDirection)
			}
		})
	}
}

// ============ Params Methods Tests ============

func TestParams_IsCursorBased(t *testing.T) {
	tests := []struct {
		name     string
		params   Params
		expected bool
	}{
		{
			name:     "no cursor",
			params:   Params{},
			expected: false,
		},
		{
			name:     "with cursor string",
			params:   Params{Cursor: "some-cursor"},
			expected: true,
		},
		{
			name:     "with parsed cursor",
			params:   Params{ParsedCursor: &Cursor{ID: 1}},
			expected: true,
		},
		{
			name:     "with both",
			params:   Params{Cursor: "cursor", ParsedCursor: &Cursor{ID: 1}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.params.IsCursorBased() != tt.expected {
				t.Errorf("IsCursorBased() = %v, want %v", tt.params.IsCursorBased(), tt.expected)
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantOffset int
	}{
		{"page 1, limit 20", 1, 20, 0},
		{"page 2, limit 20", 2, 20, 20},
		{"page 3, limit 10", 3, 10, 20},
		{"page 5, limit 50", 5, 50, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := Params{Page: tt.page, Limit: tt.limit}
			if params.Offset() != tt.wantOffset {
				t.Errorf("Offset() = %d, want %d", params.Offset(), tt.wantOffset)
			}
		})
	}
}

// ============ Result Tests ============

func TestNewResult(t *testing.T) {
	params := &Params{Page: 2, Limit: 20}
	data := []string{"item1", "item2"}
	total := 100

	result := NewResult(data, total, params)

	if result.Data == nil {
		t.Error("Data should not be nil")
	}

	if result.Total != 100 {
		t.Errorf("Total = %d, want 100", result.Total)
	}

	if result.Limit != 20 {
		t.Errorf("Limit = %d, want 20", result.Limit)
	}

	if result.Page != 2 {
		t.Errorf("Page = %d, want 2", result.Page)
	}

	if result.TotalPages != 5 {
		t.Errorf("TotalPages = %d, want 5", result.TotalPages)
	}

	if !result.HasMore {
		t.Error("HasMore should be true (page 2 of 5)")
	}
}

func TestNewResult_LastPage(t *testing.T) {
	params := &Params{Page: 5, Limit: 20}
	total := 100

	result := NewResult(nil, total, params)

	if result.HasMore {
		t.Error("HasMore should be false on last page")
	}
}

func TestNewResult_CursorBased(t *testing.T) {
	params := &Params{
		Cursor:       "some-cursor",
		ParsedCursor: &Cursor{ID: 1},
		Limit:        20,
	}

	result := NewResult(nil, 100, params)

	// For cursor-based, Page and TotalPages should be 0
	if result.Page != 0 {
		t.Errorf("Page = %d, want 0 for cursor-based", result.Page)
	}
}

func TestResult_SetCursors(t *testing.T) {
	result := &Result{}
	now := time.Now()

	result.SetCursors(true, 100, now, 1, now.Add(-time.Hour))

	if !result.HasMore {
		t.Error("HasMore should be true")
	}

	if result.NextCursor == "" {
		t.Error("NextCursor should not be empty")
	}

	if result.PrevCursor == "" {
		t.Error("PrevCursor should not be empty")
	}
}

func TestResult_SetCursors_NoMore(t *testing.T) {
	result := &Result{}

	result.SetCursors(false, 0, time.Time{}, 0, time.Time{})

	if result.HasMore {
		t.Error("HasMore should be false")
	}

	if result.NextCursor != "" {
		t.Error("NextCursor should be empty when hasMore is false")
	}
}

// ============ BuildCursorQuery Tests ============

func TestBuildCursorQuery_NilCursor(t *testing.T) {
	query := BuildCursorQuery(nil, "next")

	if query.Where != "" {
		t.Errorf("Where = %q, want empty string", query.Where)
	}

	if query.Order != "created_at DESC, id DESC" {
		t.Errorf("Order = %q, want 'created_at DESC, id DESC'", query.Order)
	}
}

func TestBuildCursorQuery_NextDirection(t *testing.T) {
	cursor := &Cursor{
		ID:        42,
		CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	query := BuildCursorQuery(cursor, "next")

	if query.Where == "" {
		t.Error("Where should not be empty")
	}

	if len(query.Args) != 3 {
		t.Errorf("Args length = %d, want 3", len(query.Args))
	}

	if query.Order != "created_at DESC, id DESC" {
		t.Errorf("Order = %q, want 'created_at DESC, id DESC'", query.Order)
	}
}

func TestBuildCursorQuery_PrevDirection(t *testing.T) {
	cursor := &Cursor{
		ID:        42,
		CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	query := BuildCursorQuery(cursor, "prev")

	if query.Where == "" {
		t.Error("Where should not be empty")
	}

	// For prev direction, order should be ASC
	if query.Order != "created_at ASC, id ASC" {
		t.Errorf("Order = %q, want 'created_at ASC, id ASC'", query.Order)
	}
}

// ============ Cursor Structure Tests ============

func TestCursor_Structure(t *testing.T) {
	cursor := Cursor{
		ID:        123,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if cursor.ID != 123 {
		t.Errorf("ID = %d, want 123", cursor.ID)
	}
}

// ============ Helper Functions ============

func createRequest(values url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/?"+values.Encode(), nil)
	return req
}
