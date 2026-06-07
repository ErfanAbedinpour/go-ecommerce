package dashboard

import (
	"strings"
	"time"
)

// ParsePeriod validates and parses a period string.
func ParsePeriod(value string) (Period, error) {
	switch Period(strings.ToLower(strings.TrimSpace(value))) {
	case PeriodToday, PeriodWeek, PeriodMonth, PeriodYear:
		return Period(strings.ToLower(strings.TrimSpace(value))), nil
	case "":
		return PeriodMonth, nil
	default:
		return "", ErrInvalidPeriod
	}
}

// ResolveDateRange converts a revenue filter into a concrete query window.
func ResolveDateRange(filter RevenueFilter) (DateRange, error) {
	now := time.Now().UTC()

	if filter.From != nil && filter.To != nil {
		from := filter.From.UTC()
		to := filter.To.UTC()
		if from.After(to) {
			return DateRange{}, ErrInvalidDateRange
		}
		// Include the entire end day.
		to = to.Add(24*time.Hour - time.Nanosecond)
		return DateRange{
			From:        from,
			To:          to,
			Granularity: granularityForDuration(to.Sub(from)),
		}, nil
	}

	period, err := ParsePeriod(string(filter.Period))
	if err != nil {
		return DateRange{}, err
	}

	switch period {
	case PeriodToday:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return DateRange{From: start, To: now, Granularity: GranularityHour}, nil
	case PeriodWeek:
		start := now.AddDate(0, 0, -6)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		return DateRange{From: start, To: now, Granularity: GranularityDay}, nil
	case PeriodMonth:
		start := now.AddDate(0, 0, -29)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		return DateRange{From: start, To: now, Granularity: GranularityDay}, nil
	case PeriodYear:
		start := now.AddDate(-1, 0, 0)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
		return DateRange{From: start, To: now, Granularity: GranularityMonth}, nil
	default:
		return DateRange{}, ErrInvalidPeriod
	}
}

func granularityForDuration(d time.Duration) Granularity {
	switch {
	case d <= 48*time.Hour:
		return GranularityHour
	case d <= 62*24*time.Hour:
		return GranularityDay
	default:
		return GranularityMonth
	}
}
