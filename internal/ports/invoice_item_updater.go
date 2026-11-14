package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemUpdater interface {
	Update(tenantID string, item *invoiceitem.InvoiceItem) error
}
