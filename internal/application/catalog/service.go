// Package catalog expone los casos de uso mínimos de clientes y productos:
// crear y buscar dentro del tenant autenticado.
package catalog

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type Service struct {
	customers ports.CustomerRepository
	products  ports.ProductRepository
}

func NewService(customers ports.CustomerRepository, products ports.ProductRepository) *Service {
	return &Service{customers: customers, products: products}
}

func (s *Service) CreateCustomer(tenantID, name, document string) (*catalog.Customer, error) {
	c, err := catalog.NewCustomer(tenantID, name, document)
	if err != nil {
		return nil, err
	}
	if err := s.customers.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListCustomers(tenantID string, q ports.CatalogQuery) ([]*catalog.Customer, int, error) {
	return s.customers.List(tenantID, q)
}

func (s *Service) CreateProduct(tenantID, code, name string, unitPrice float64) (*catalog.Product, error) {
	p, err := catalog.NewProduct(tenantID, code, name, unitPrice)
	if err != nil {
		return nil, err
	}
	if err := s.products.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListProducts(tenantID string, q ports.CatalogQuery) ([]*catalog.Product, int, error) {
	return s.products.List(tenantID, q)
}
