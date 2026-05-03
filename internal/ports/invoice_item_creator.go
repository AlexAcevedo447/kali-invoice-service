package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemCreator interface {
	Save(item *invoiceitem.InvoiceItem) error
}
