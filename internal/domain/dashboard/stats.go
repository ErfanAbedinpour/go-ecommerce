package dashboard

// Stats holds aggregated dashboard KPIs.
type Stats struct {
	TotalRevenue   float64
	TotalOrders    int64
	TotalCustomers int64
	TotalProducts  int64
	PendingOrders  int64
	LowStockCount  int64
}
