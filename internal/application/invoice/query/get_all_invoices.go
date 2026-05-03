package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type GetAllInvoicesQuery struct {
	repo ports.InvoiceListReader
}

func NewGetAllInvoicesQuery(repo ports.InvoiceListReader) *GetAllInvoicesQuery {
	return &GetAllInvoicesQuery{repo: repo}
}

func (q *GetAllInvoicesQuery) Execute(p ports.Pagination) ([]*invoice.Invoice, error) {
	return q.repo.GetAll(p)
}
