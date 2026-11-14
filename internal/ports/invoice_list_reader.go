package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceListReader consulta colecciones paginadas.
type InvoiceListReader interface {
	GetAll(tenantID string, p Pagination) ([]*invoice.Invoice, error)
	// Count devuelve el total de facturas, para que el cliente pueda calcular páginas.
	Count(tenantID string) (int, error)
}
