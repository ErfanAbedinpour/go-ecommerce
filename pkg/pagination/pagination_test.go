package pagination

import (
	"net/http/httptest"
	"testing"
)

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
		wantOrder   string
	}{
		{"defaults", "", DefaultPage, DefaultPerPage, "desc"},
		{"custom page", "?page=3&per_page=50", 3, 50, "desc"},
		{"max per page capped", "?per_page=200", DefaultPage, MaxPerPage, "desc"},
		{"asc order", "?order=asc", DefaultPage, DefaultPerPage, "asc"},
		{"invalid page", "?page=0", DefaultPage, DefaultPerPage, "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/"+tt.query, nil)
			got := FromRequest(r)

			if got.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", got.Page, tt.wantPage)
			}
			if got.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %d, want %d", got.PerPage, tt.wantPerPage)
			}
			if got.Order != tt.wantOrder {
				t.Errorf("Order = %q, want %q", got.Order, tt.wantOrder)
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	p := Params{Page: 3, PerPage: 20}
	if got := p.Offset(); got != 40 {
		t.Errorf("Offset() = %d, want 40", got)
	}
}

func TestNewMeta(t *testing.T) {
	meta := NewMeta(1, 20, 45)
	if meta.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", meta.TotalPages)
	}
	if meta.Total != 45 {
		t.Errorf("Total = %d, want 45", meta.Total)
	}
}

func TestNewPaginated_EmptySlice(t *testing.T) {
	result := NewPaginated([]string{}, 1, 20, 0)
	if result.Data == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(result.Data) != 0 {
		t.Errorf("expected empty data, got %d items", len(result.Data))
	}
}
