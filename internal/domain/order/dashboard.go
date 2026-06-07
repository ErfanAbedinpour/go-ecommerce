package order

// DashboardSummary is a read-model for the dashboard recent-orders feed.
type DashboardSummary struct {
	Summary
	CustomerName string
	ProductName  string
}
