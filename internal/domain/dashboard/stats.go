package dashboard

// Stats holds aggregated dashboard KPIs.
type Stats struct {
	TotalRevenue   float64
	TotalOrders    int64
	TotalCustomers int64
	TotalProducts  int64
	PendingOrders  int64
	LowStockCount  int64
	Growth         StatsGrowth
}

// StatsGrowth holds period-over-period growth percentages (last 30 days vs prior 30 days).
type StatsGrowth struct {
	TotalRevenue   float64
	TotalOrders    float64
	TotalCustomers float64
	TotalProducts  float64
	PendingOrders  float64
	LowStockCount  float64
}
