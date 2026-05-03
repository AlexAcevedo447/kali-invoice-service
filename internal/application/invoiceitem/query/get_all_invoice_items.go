package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type GetAllInvoiceItemsQuery struct {
	repo ports.InvoiceItemListReader
}

func NewGetAllInvoiceItemsQuery(repo ports.InvoiceItemListReader) *GetAllInvoiceItemsQuery {
	return &GetAllInvoiceItemsQuery{repo: repo}
}

func (q *GetAllInvoiceItemsQuery) Execute(p ports.Pagination) ([]*invoiceitem.InvoiceItem, error) {
	return q.repo.GetAll(p)
}
