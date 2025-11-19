package query

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/contracts"
)

type AppGetAllInvoicesQuery struct{
	Repo contracts.IGetAllInvoiceQueryRepository
}

func NewAppGetAllInvoicesQuery(repo contracts.IGetAllInvoiceQueryRepository) *AppGetAllInvoicesQuery {
	return &AppGetAllInvoicesQuery{
		Repo: repo,
	}
}

func (query *AppGetAllInvoicesQuery) Execute()([]*entities.Invoice, error){
	invoice, getErr := query.Repo.GetAll()

	if getErr != nil {
		return nil, getErr
	}

	return invoice, nil
}