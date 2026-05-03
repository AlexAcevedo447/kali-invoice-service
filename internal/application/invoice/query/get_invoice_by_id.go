package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

type GetInvoiceByIDQuery struct {
	repo ports.InvoiceByIDReader
}

func NewGetInvoiceByIDQuery(repo ports.InvoiceByIDReader) *GetInvoiceByIDQuery {
	return &GetInvoiceByIDQuery{repo: repo}
}

func (q *GetInvoiceByIDQuery) Execute(id string) (*invoice.Invoice, error) {
	return q.repo.GetByID(id)
}
