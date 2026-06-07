package dashboard

import "time"

// Period identifies a preset revenue analytics window.
type Period string

const (
	PeriodToday Period = "today"
	PeriodWeek  Period = "week"
	PeriodMonth Period = "month"
	PeriodYear  Period = "year"
)

// Granularity controls how revenue data points are grouped.
type Granularity string

const (
	GranularityHour  Granularity = "hour"
	GranularityDay   Granularity = "day"
	GranularityMonth Granularity = "month"
)

// RevenueFilter holds parameters for revenue analytics queries.
type RevenueFilter struct {
	Period Period
	From   *time.Time
	To     *time.Time
}

// DateRange is a resolved inclusive-exclusive analytics window.
type DateRange struct {
	From        time.Time
	To          time.Time
	Granularity Granularity
}

// RevenueDataPoint is a single time-series revenue entry.
type RevenueDataPoint struct {
	Date    string
	Revenue float64
	Orders  int64
}
