package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"

// InvoiceCreator persiste facturas nuevas.
type InvoiceCreator interface {
	Save(inv *invoice.Invoice) error
}
