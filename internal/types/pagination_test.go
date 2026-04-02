package types_test

import (
	"testing"

	"github.com/larsartmann/complaints-mcp/internal/types"
)

func assertPageRequestFields(t *testing.T, page, perPage, wantPage, wantPerPage int) {
	got := types.NewPageRequest(page, perPage)
	if got.Page != wantPage {
		t.Errorf("Page = %d, want %d", got.Page, wantPage)
	}

	if got.PerPage != wantPerPage {
		t.Errorf("PerPage = %d, want %d", got.PerPage, wantPerPage)
	}
}

func assertCursorRequestFields(t *testing.T, cursor string, limit, wantLimit int) {
	got := types.NewCursorRequest(cursor, limit)
	if got.Cursor != cursor {
		t.Errorf("Cursor = %q, want %q", got.Cursor, cursor)
	}

	if got.Limit != wantLimit {
		t.Errorf("Limit = %d, want %d", got.Limit, wantLimit)
	}
}

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		perPage     int
		wantPage    int
		wantPerPage int
	}{
		{
			name:        "valid values",
			page:        2,
			perPage:     50,
			wantPage:    2,
			wantPerPage: 50,
		},
		{
			name:        "zero page defaults to 1",
			page:        0,
			perPage:     20,
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "negative page defaults to 1",
			page:        -1,
			perPage:     20,
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "zero per_page defaults to 20",
			page:        1,
			perPage:     0,
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "negative per_page defaults to 20",
			page:        1,
			perPage:     -5,
			wantPage:    1,
			wantPerPage: 20,
		},
		{
			name:        "per_page capped at max",
			page:        1,
			perPage:     200,
			wantPage:    1,
			wantPerPage: 100,
		},
		{
			name:        "all defaults",
			page:        0,
			perPage:     0,
			wantPage:    1,
			wantPerPage: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPageRequestFields(t, tt.page, tt.perPage, tt.wantPage, tt.wantPerPage)
		})
	}
}

func TestPageRequest_Offset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		wantOffset int
	}{
		{
			name:       "page 1",
			page:       1,
			perPage:    20,
			wantOffset: 0,
		},
		{
			name:       "page 2",
			page:       2,
			perPage:    20,
			wantOffset: 20,
		},
		{
			name:       "page 5 with 50 per page",
			page:       5,
			perPage:    50,
			wantOffset: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := types.NewPageRequest(tt.page, tt.perPage)
			if got := pr.Offset(); got != tt.wantOffset {
				t.Errorf("Offset() = %d, want %d", got, tt.wantOffset)
			}
		})
	}
}

func TestPageRequest_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		perPage int
		want    bool
	}{
		{
			name:    "valid",
			page:    1,
			perPage: 20,
			want:    true,
		},
		{
			name:    "invalid page",
			page:    0,
			perPage: 20,
			want:    false,
		},
		{
			name:    "invalid per_page too low",
			page:    1,
			perPage: 0,
			want:    false,
		},
		{
			name:    "invalid per_page too high",
			page:    1,
			perPage: 101,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := types.PageRequest{Page: tt.page, PerPage: tt.perPage}
			if got := pr.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPageResponse(t *testing.T) {
	tests := []struct {
		name           string
		data           []string
		req            types.PageRequest
		totalItems     int
		wantPage       int
		wantTotalPages int
		wantHasNext    bool
		wantHasPrev    bool
	}{
		{
			name:           "first page with more items",
			data:           []string{"a", "b", "c"},
			req:            types.NewPageRequest(1, 10),
			totalItems:     25,
			wantPage:       1,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    false,
		},
		{
			name:           "last page",
			data:           []string{"x", "y"},
			req:            types.NewPageRequest(3, 10),
			totalItems:     22,
			wantPage:       3,
			wantTotalPages: 3,
			wantHasNext:    false,
			wantHasPrev:    true,
		},
		{
			name:           "middle page",
			data:           []string{"m"},
			req:            types.NewPageRequest(2, 10),
			totalItems:     21,
			wantPage:       2,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    true,
		},
		{
			name:           "empty result",
			data:           []string{},
			req:            types.NewPageRequest(1, 10),
			totalItems:     0,
			wantPage:       1,
			wantTotalPages: 1,
			wantHasNext:    false,
			wantHasPrev:    false,
		},
		{
			name:           "exact page boundary",
			data:           []string{"a", "b"},
			req:            types.NewPageRequest(1, 2),
			totalItems:     4,
			wantPage:       1,
			wantTotalPages: 2,
			wantHasNext:    true,
			wantHasPrev:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.NewPageResponse(tt.data, tt.req, tt.totalItems)

			if got.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", got.Page, tt.wantPage)
			}

			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}

			if got.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", got.HasNext, tt.wantHasNext)
			}

			if got.HasPrev != tt.wantHasPrev {
				t.Errorf("HasPrev = %v, want %v", got.HasPrev, tt.wantHasPrev)
			}

			if got.TotalItems != tt.totalItems {
				t.Errorf("TotalItems = %d, want %d", got.TotalItems, tt.totalItems)
			}

			if len(got.Data) != len(tt.data) {
				t.Errorf("len(Data) = %d, want %d", len(got.Data), len(tt.data))
			}
		})
	}
}

func TestPageResponse_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		data []string
		want bool
	}{
		{
			name: "not empty",
			data: []string{"a"},
			want: false,
		},
		{
			name: "empty",
			data: []string{},
			want: true,
		},
		{
			name: "nil",
			data: nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := types.PageResponse[string]{Data: tt.data}
			if got := pr.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmptyPageResponse(t *testing.T) {
	req := types.NewPageRequest(2, 20)
	resp := types.EmptyPageResponse[string](req)

	if !resp.IsEmpty() {
		t.Error("EmptyPageResponse should return empty response")
	}

	if resp.Page != 2 {
		t.Errorf("Page = %d, want 2", resp.Page)
	}

	if resp.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", resp.PerPage)
	}

	if resp.HasPrev != true {
		t.Error("HasPrev should be true for page 2")
	}

	if resp.HasNext != false {
		t.Error("HasNext should be false")
	}
}

func TestNewCursorRequest(t *testing.T) {
	tests := []struct {
		name      string
		cursor    string
		limit     int
		wantLimit int
	}{
		{
			name:      "valid values",
			cursor:    "abc123",
			limit:     50,
			wantLimit: 50,
		},
		{
			name:      "empty cursor",
			cursor:    "",
			limit:     20,
			wantLimit: 20,
		},
		{
			name:      "zero limit defaults to 20",
			cursor:    "xyz",
			limit:     0,
			wantLimit: 20,
		},
		{
			name:      "negative limit defaults to 20",
			cursor:    "xyz",
			limit:     -5,
			wantLimit: 20,
		},
		{
			name:      "limit capped at max",
			cursor:    "xyz",
			limit:     200,
			wantLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCursorRequestFields(t, tt.cursor, tt.limit, tt.wantLimit)
		})
	}
}

func TestPaginationError(t *testing.T) {
	err := types.PaginationError{
		Field:   "page",
		Message: "must be positive",
	}

	want := "pagination error for page: must be positive"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
