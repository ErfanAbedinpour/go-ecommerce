package shared

// PageParams holds domain-level pagination parameters.
type PageParams struct {
	Page    int
	PerPage int
	Sort    string
	Order   string
}

// PageResult holds paginated query results.
type PageResult[T any] struct {
	Items []T
	Total int64
}
