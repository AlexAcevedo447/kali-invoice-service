package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemByIDReader interface {
	GetByID(tenantID, id string) (*invoiceitem.InvoiceItem, error)
}
