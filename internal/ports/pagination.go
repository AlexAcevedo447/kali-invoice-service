package ports

// Pagination define los parámetros de paginación para consultas de colección.
type Pagination struct {
	Page     int // 1-based
	PageSize int
}

// DefaultPagination retorna la paginación por defecto (página 1, 20 registros).
func DefaultPagination() Pagination {
	return Pagination{Page: 1, PageSize: 20}
}
