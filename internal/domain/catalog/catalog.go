// Package catalog contiene los catálogos mínimos de facturación por tenant:
// clientes y productos. Permiten que el usuario elija entidades legibles
// mientras el sistema guarda internamente su identificador (UUID).
package catalog

import (
	"errors"
	"strings"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/domainerr"
	"github.com/google/uuid"
)

var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrProductNotFound       = errors.New("product not found")
	ErrCustomerNameRequired  = domainerr.NewValidation("customer name is required")
	ErrCustomerDocRequired   = domainerr.NewValidation("customer document is required")
	ErrCustomerDocDuplicated = errors.New("a customer with this document already exists")
	ErrProductCodeRequired   = domainerr.NewValidation("product code is required")
	ErrProductNameRequired   = domainerr.NewValidation("product name is required")
	ErrProductPriceInvalid   = domainerr.NewValidation("product unit_price must be greater than or equal to 0")
	ErrProductCodeDuplicated = errors.New("a product with this code already exists")
	ErrTenantRequired        = errors.New("tenant is required")
)

// Customer es un cliente de la empresa (tenant).
type Customer struct {
	ID        string
	TenantID  string
	Name      string
	Document  string
	CreatedAt time.Time
}

// Product es un producto o servicio facturable de la empresa (tenant).
type Product struct {
	ID        string
	TenantID  string
	Code      string
	Name      string
	UnitPrice float64
	CreatedAt time.Time
}

// NewCustomer valida y construye un cliente con UUID generado en el servidor.
func NewCustomer(tenantID, name, document string) (*Customer, error) {
	name, document = strings.TrimSpace(name), strings.TrimSpace(document)
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if name == "" {
		return nil, ErrCustomerNameRequired
	}
	if document == "" {
		return nil, ErrCustomerDocRequired
	}
	return &Customer{ID: uuid.New().String(), TenantID: tenantID, Name: name, Document: document, CreatedAt: time.Now()}, nil
}

// NewProduct valida y construye un producto con UUID generado en el servidor.
func NewProduct(tenantID, code, name string, unitPrice float64) (*Product, error) {
	code, name = strings.TrimSpace(code), strings.TrimSpace(name)
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	if code == "" {
		return nil, ErrProductCodeRequired
	}
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if unitPrice < 0 {
		return nil, ErrProductPriceInvalid
	}
	return &Product{ID: uuid.New().String(), TenantID: tenantID, Code: code, Name: name, UnitPrice: unitPrice, CreatedAt: time.Now()}, nil
}
