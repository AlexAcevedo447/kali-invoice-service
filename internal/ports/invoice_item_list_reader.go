package ports

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

type InvoiceItemListReader interface {
	GetAll(p Pagination) ([]*invoiceitem.InvoiceItem, error)
}