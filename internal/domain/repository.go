package domain

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"

type CancelInvoiceCommandRepository interface {
	Cancel(id string) error
}

type CreateInvoiceCommandRepository interface {
	Save(invoice *entities.Invoice) error
}

type UpdateInvoiceCommandRepository interface {
	Update(invoice *entities.Invoice) error
}

type GetByIdInvoiceQueryRepository interface {
	GetById(id string) (*entities.Invoice, error)
}

type GetAllInvoiceQueryRepository interface {
	GetAll() ([]*entities.Invoice, error)
}