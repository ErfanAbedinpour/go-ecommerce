package dashboard

import (
	"testing"
	"time"
)

func TestParsePeriod_DefaultMonth(t *testing.T) {
	p, err := ParsePeriod("")
	if err != nil {
		t.Fatalf("ParsePeriod() error = %v", err)
	}
	if p != PeriodMonth {
		t.Errorf("period = %q, want month", p)
	}
}

func TestParsePeriod_Invalid(t *testing.T) {
	_, err := ParsePeriod("invalid")
	if err != ErrInvalidPeriod {
		t.Errorf("expected invalid period, got %v", err)
	}
}

func TestResolveDateRange_CustomRange(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	range_, err := ResolveDateRange(RevenueFilter{From: &from, To: &to})
	if err != nil {
		t.Fatalf("ResolveDateRange() error = %v", err)
	}
	if range_.Granularity != GranularityDay {
		t.Errorf("granularity = %q, want day", range_.Granularity)
	}
}

func TestResolveDateRange_InvalidRange(t *testing.T) {
	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := ResolveDateRange(RevenueFilter{From: &from, To: &to})
	if err != ErrInvalidDateRange {
		t.Errorf("expected invalid date range, got %v", err)
	}
}

func TestResolveDateRange_Today(t *testing.T) {
	range_, err := ResolveDateRange(RevenueFilter{Period: PeriodToday})
	if err != nil {
		t.Fatalf("ResolveDateRange() error = %v", err)
	}
	if range_.Granularity != GranularityHour {
		t.Errorf("granularity = %q, want hour", range_.Granularity)
	}
}
