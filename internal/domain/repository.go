package domain

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"

type ICancelInvoiceCommandRepository interface {
	Cancel(id string) error
}

type ICreateInvoiceCommandRepository interface {
	Save(invoice *entities.Invoice) error
}

type IUpdateInvoiceCommandRepository interface {
	Update(invoice *entities.Invoice) error
}

type IGetByIdInvoiceQueryRepository interface {
	GetById(id string) (*entities.Invoice, error)
}

type IGetAllInvoiceQueryRepository interface {
	GetAll() ([]*entities.Invoice, error)
}