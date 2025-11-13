package commands

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
)

type CreateInvoiceCommand struct {
	repo domain.ICreateInvoiceCommandRepository
}

func NewCreateInvoiceCommand(repo domain.ICreateInvoiceCommandRepository) *CreateInvoiceCommand {
	return &CreateInvoiceCommand{repo: repo}
}

func (cmd *CreateInvoiceCommand) Execute(dto entities.NewInvoiceDTO)(*entities.Invoice, error){
	invoice, err := entities.NewInvoice(dto)

	if err != nil {
		return nil, err
	}

	if err:= invoice.Validate(); err != nil {
		return nil, err
	}

	if err := cmd.repo.Save(invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}