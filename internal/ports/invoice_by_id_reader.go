package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceByIDReader consulta una factura puntual.
type InvoiceByIDReader interface {
	GetByID(id string) (*invoice.Invoice, error)
}