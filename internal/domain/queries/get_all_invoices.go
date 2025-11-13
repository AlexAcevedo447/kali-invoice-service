package queries

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
)

type GetAllInvoicesQuery struct {
	repo domain.IGetAllInvoiceQueryRepository
}

func NewGetAllInvoicesQuery(repo domain.IGetByIdInvoiceQueryRepository) *GetInvoiceByIdQuery {
	return &GetInvoiceByIdQuery{repo: repo}
}

func (q *GetAllInvoicesQuery) Execute() ([]*entities.Invoice, error) {
	return q.repo.GetAll()
}
