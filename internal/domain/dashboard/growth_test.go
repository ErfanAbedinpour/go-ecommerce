package dashboard

import "testing"

func TestCalcGrowthPercent(t *testing.T) {
	tests := []struct {
		current  float64
		previous float64
		want     float64
	}{
		{110, 100, 10},
		{90, 100, -10},
		{50, 0, 100},
		{0, 0, 0},
		{12.55, 10, 25.5},
	}

	for _, tt := range tests {
		got := CalcGrowthPercent(tt.current, tt.previous)
		if got != tt.want {
			t.Errorf("CalcGrowthPercent(%v, %v) = %v, want %v", tt.current, tt.previous, got, tt.want)
		}
	}
}
