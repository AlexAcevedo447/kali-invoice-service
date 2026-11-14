package mappers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
)

func CustomerToModel(c *catalog.Customer) *models.CustomerModel {
	return &models.CustomerModel{ID: c.ID, TenantID: c.TenantID, Name: c.Name, Document: c.Document, CreatedAt: c.CreatedAt}
}

func ModelToCustomer(m *models.CustomerModel) *catalog.Customer {
	return &catalog.Customer{ID: m.ID, TenantID: m.TenantID, Name: m.Name, Document: m.Document, CreatedAt: m.CreatedAt}
}

func ProductToModel(p *catalog.Product) *models.ProductModel {
	return &models.ProductModel{ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, UnitPrice: p.UnitPrice, CreatedAt: p.CreatedAt}
}

func ModelToProduct(m *models.ProductModel) *catalog.Product {
	return &catalog.Product{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, UnitPrice: m.UnitPrice, CreatedAt: m.CreatedAt}
}
