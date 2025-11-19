package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/dto"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/contracts"
)

type AppGetByIdInvoiceQuery struct{
	Repo contracts.IGetByIdInvoiceQueryRepository
}

func NewAppGetByIdInvoiceQuery(repo contracts.IGetByIdInvoiceQueryRepository) *AppGetByIdInvoiceQuery {
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