package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

// ============ Constants Tests ============

func TestPaginationConstants(t *testing.T) {
	if DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want 1", DefaultPage)
	}
	if DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d, want 20", DefaultLimit)
	}
	if MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want 100", MaxLimit)
	}
}

// ============ ParsePagination Tests ============

func TestParsePagination_Defaults(t *testing.T) {
	req := createPaginationRequest(url.Values{})

	p := ParsePagination(req)

	if p.Page != DefaultPage {
		t.Errorf("Page = %d, want %d", p.Page, DefaultPage)
	}
	if p.Limit != DefaultLimit {
		t.Errorf("Limit = %d, want %d", p.Limit, DefaultLimit)
	}
	if p.Offset != 0 {
		t.Errorf("Offset = %d, want 0", p.Offset)
	}
}

func TestParsePagination_ValidValues(t *testing.T) {
	tests := []struct {
		name       string
		page       string
		limit      string
		wantPage   int
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "page 1 limit 10",
			page:       "1",
			limit:      "10",
			wantPage:   1,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "page 2 limit 20",
			page:       "2",
			limit:      "20",
			wantPage:   2,
			wantLimit:  20,
			wantOffset: 20,
		},
		{
			name:       "page 3 limit 50",
			page:       "3",
			limit:      "50",
			wantPage:   3,
			wantLimit:  50,
			wantOffset: 100,
		},
		{
			name:       "page 5 limit 25",
			page:       "5",
			limit:      "25",
			wantPage:   5,
			wantLimit:  25,
			wantOffset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			values.Set("page", tt.page)
			values.Set("limit", tt.limit)
			req := createPaginationRequest(values)

			p := ParsePagination(req)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
			}
			if p.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", p.Offset, tt.wantOffset)
			}
		})
	}
}

func TestParsePagination_InvalidPage(t *testing.T) {
	tests := []struct {
		name     string
		page     string
		wantPage int
	}{
		{"zero page", "0", DefaultPage},
		{"negative page", "-1", DefaultPage},
		{"non-numeric page", "abc", DefaultPage},
		{"empty page", "", DefaultPage},
		{"float page", "1.5", DefaultPage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			values.Set("page", tt.page)
			req := createPaginationRequest(values)

			p := ParsePagination(req)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
		})
	}
}

func TestParsePagination_InvalidLimit(t *testing.T) {
	tests := []struct {
		name      string
		limit     string
		wantLimit int
	}{
		{"zero limit", "0", DefaultLimit},
		{"negative limit", "-1", DefaultLimit},
		{"non-numeric limit", "abc", DefaultLimit},
		{"empty limit", "", DefaultLimit},
		{"float limit", "10.5", DefaultLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			values.Set("limit", tt.limit)
			req := createPaginationRequest(values)

			p := ParsePagination(req)

			if p.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
			}
		})
	}
}

func TestParsePagination_LimitCapping(t *testing.T) {
	tests := []struct {
		name      string
		limit     string
		wantLimit int
	}{
		{"at max limit", "100", 100},
		{"over max limit", "150", DefaultLimit},
		{"way over max limit", "1000", DefaultLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			values.Set("limit", tt.limit)
			req := createPaginationRequest(values)

			p := ParsePagination(req)

			if p.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
			}
		})
	}
}

// ============ ParsePaginationWithDefaults Tests ============

func TestParsePaginationWithDefaults_CustomValues(t *testing.T) {
	tests := []struct {
		name         string
		defaultLimit int
		maxLimit     int
		pageParam    string
		limitParam   string
		wantPage     int
		wantLimit    int
		wantOffset   int
	}{
		{
			name:         "custom defaults applied",
			defaultLimit: 10,
			maxLimit:     50,
			pageParam:    "",
			limitParam:   "",
			wantPage:     1,
			wantLimit:    10,
			wantOffset:   0,
		},
		{
			name:         "custom max limit enforced",
			defaultLimit: 10,
			maxLimit:     50,
			pageParam:    "1",
			limitParam:   "100",
			wantPage:     1,
			wantLimit:    10, // Over max, falls back to default
			wantOffset:   0,
		},
		{
			name:         "valid limit within custom max",
			defaultLimit: 10,
			maxLimit:     50,
			pageParam:    "2",
			limitParam:   "30",
			wantPage:     2,
			wantLimit:    30,
			wantOffset:   30,
		},
		{
			name:         "at custom max limit",
			defaultLimit: 10,
			maxLimit:     50,
			pageParam:    "1",
			limitParam:   "50",
			wantPage:     1,
			wantLimit:    50,
			wantOffset:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			if tt.pageParam != "" {
				values.Set("page", tt.pageParam)
			}
			if tt.limitParam != "" {
				values.Set("limit", tt.limitParam)
			}
			req := createPaginationRequest(values)

			p := ParsePaginationWithDefaults(req, tt.defaultLimit, tt.maxLimit)

			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
			}
			if p.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", p.Offset, tt.wantOffset)
			}
		})
	}
}

// ============ Pagination Structure Tests ============

func TestPagination_Structure(t *testing.T) {
	p := Pagination{
		Page:   5,
		Limit:  25,
		Offset: 100,
	}

	if p.Page != 5 {
		t.Errorf("Page = %d, want 5", p.Page)
	}
	if p.Limit != 25 {
		t.Errorf("Limit = %d, want 25", p.Limit)
	}
	if p.Offset != 100 {
		t.Errorf("Offset = %d, want 100", p.Offset)
	}
}

// ============ Offset Calculation Tests ============

func TestParsePagination_OffsetCalculation(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantOffset int
	}{
		{"page 1", 1, 20, 0},
		{"page 2", 2, 20, 20},
		{"page 10", 10, 20, 180},
		{"page 1 limit 50", 1, 50, 0},
		{"page 3 limit 50", 3, 50, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := url.Values{}
			values.Set("page", strconv.Itoa(tt.page))
			values.Set("limit", strconv.Itoa(tt.limit))
			req := createPaginationRequest(values)

			p := ParsePagination(req)

			if p.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", p.Offset, tt.wantOffset)
			}
		})
	}
}

// ============ Helper Functions ============

func createPaginationRequest(values url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/?"+values.Encode(), nil)
	return req
}
