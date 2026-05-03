package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type GetInvoiceItemByIDQuery struct {
	repo ports.InvoiceItemByIDReader
}

func NewGetInvoiceItemByIDQuery(repo ports.InvoiceItemByIDReader) *GetInvoiceItemByIDQuery {
	return &GetInvoiceItemByIDQuery{repo: repo}
}

func (q *GetInvoiceItemByIDQuery) Execute(id string) (*invoiceitem.InvoiceItem, error) {
	return q.repo.GetByID(id)
}