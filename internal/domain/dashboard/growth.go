package dashboard

import "math"

// CalcGrowthPercent returns the percentage change from previous to current, rounded to one decimal.
func CalcGrowthPercent(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	growth := ((current - previous) / previous) * 100
	return math.Round(growth*10) / 10
}
