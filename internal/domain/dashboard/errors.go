package dashboard

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrInvalidPeriod = apperror.ValidationKeyed(i18n.KeyDashboardInvalidPeriod, "invalid period", map[string]string{
		"period": "must be one of: today, week, month, year",
	})
	ErrInvalidDateRange = apperror.ValidationKeyed(i18n.KeyDashboardInvalidDateRange, "invalid date range", map[string]string{
		"date_range": "start date must be before end date",
	})
	ErrInvalidLimit = apperror.ValidationKeyed(i18n.KeyDashboardInvalidLimit, "invalid limit", map[string]string{
		"limit": "must be between 1 and 100",
	})
)
