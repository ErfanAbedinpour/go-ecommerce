package product

// Stats holds product catalog KPI counts.
type Stats struct {
	Total      int64 `gorm:"column:total"`
	Active     int64 `gorm:"column:active"`
	Draft      int64 `gorm:"column:draft"`
	OutOfStock int64 `gorm:"column:out_of_stock"`
}
