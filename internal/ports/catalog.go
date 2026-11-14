package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"

// CatalogQuery filtra los catálogos por texto legible (nombre, documento o código).
type CatalogQuery struct {
	Search     string
	Pagination Pagination
}

// CustomerRepository persiste y consulta clientes, siempre dentro de un tenant.
type CustomerRepository interface {
	Create(c *catalog.Customer) error
	List(tenantID string, q CatalogQuery) ([]*catalog.Customer, int, error)
	Exists(tenantID, id string) (bool, error)
}

// ProductRepository persiste y consulta productos, siempre dentro de un tenant.
type ProductRepository interface {
	Create(p *catalog.Product) error
	List(tenantID string, q CatalogQuery) ([]*catalog.Product, int, error)
	// MissingIDs devuelve los ids que no existen como productos del tenant.
	MissingIDs(tenantID string, ids []string) ([]string, error)
}
