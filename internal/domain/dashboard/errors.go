package dashboard

import "app/pkg/apperror"

var (
	ErrInvalidPeriod = apperror.Validation("invalid period", map[string]string{
		"period": "must be one of: today, week, month, year",
	})
	ErrInvalidDateRange = apperror.Validation("invalid date range", map[string]string{
		"from": "must be before or equal to to",
	})
	ErrInvalidLimit = apperror.Validation("invalid limit", map[string]string{
		"limit": "must be between 1 and 50",
	})
)
