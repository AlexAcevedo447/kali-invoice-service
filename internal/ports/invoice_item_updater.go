package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemUpdater interface {
	Update(item *invoiceitem.InvoiceItem) error
}
