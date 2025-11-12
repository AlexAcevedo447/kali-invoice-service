package command

import (
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/dto"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/shared/utils/helpers"
)

type AppCreateInvoiceCommand struct{
	Repo domain.CreateInvoiceCommandRepository
}

func NewAppCreateInvoiceCommand(repo domain.CreateInvoiceCommandRepository) *AppCreateInvoiceCommand {
	return &AppCreateInvoiceCommand{
		Repo: repo,
	}
}

func (command *AppCreateInvoiceCommand) Execute(dto dto.AppCreateInvoiceDTO)(*entities.Invoice, error){
	newInvoiceDto := entities.NewInvoiceDTO{
		CustomerID: dto.CustomerId,
		IssueDate: helpers.OrTime(dto.IssueDate, time.Now()),
	}

	invoice, creationErr := entities.NewInvoice(newInvoiceDto)

	if creationErr != nil {
		return nil, creationErr
	}

	err := command.Repo.Save(invoice);

	if err != nil {
		return nil, err
	}

	return invoice, nil
}