package queries

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
)

type GetInvoiceByIdQuery struct {
	repo domain.IGetByIdInvoiceQueryRepository
}

func NewGetInvoiceByIdQuery(repo domain.IGetByIdInvoiceQueryRepository) *GetInvoiceByIdQuery {
	return &GetInvoiceByIdQuery{repo: repo}
}

func (q *GetInvoiceByIdQuery) Execute(id string) (*entities.Invoice, error) {
	return q.repo.GetById(id)
}
