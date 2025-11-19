package domain

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"

type ICancelInvoiceCommandFactory interface {
	Cancel(id string) error
}

type ICreateInvoiceCommandFactory interface {
	Execute(invoice entities.NewInvoiceDTO) (*entities.Invoice, error)
}

type IUpdateInvoiceCommandFactory interface {
	Update(invoice *entities.Invoice) error
}