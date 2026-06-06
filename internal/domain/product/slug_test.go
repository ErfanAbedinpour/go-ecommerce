package product

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Nike Air Max Black", "nike-air-max-black"},
		{"  Hello   World  ", "hello-world"},
		{"Product #1 (Special!)", "product-1-special"},
		{"", "product"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateSlug(tt.name); got != tt.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
