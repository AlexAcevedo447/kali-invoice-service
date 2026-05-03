package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceListReader consulta colecciones paginadas.
type InvoiceListReader interface {
	GetAll(p Pagination) ([]*invoice.Invoice, error)
}
