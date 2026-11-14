package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemCreator interface {
	Save(tenantID string, item *invoiceitem.InvoiceItem) error
}
