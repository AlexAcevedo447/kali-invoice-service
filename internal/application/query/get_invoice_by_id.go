package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/dto"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
)

type AppGetByIdInvoiceQuery struct{
	Repo domain.IGetByIdInvoiceQueryRepository
}

func NewAppGetByIdInvoiceQuery(repo domain.IGetByIdInvoiceQueryRepository) *AppGetByIdInvoiceQuery {
	return &AppGetByIdInvoiceQuery{
		Repo: repo,
	}
}

func (query *AppGetByIdInvoiceQuery) Execute(dto dto.AppGetByIdInvoiceDTO)(*entities.Invoice, error){
	invoice, getErr := query.Repo.GetById(dto.Id)

	if getErr != nil {
		return nil, getErr
	}

	return invoice, nil
}