package domain

type InvoiceCommandRepository interface {
	Save(invoice *Invoice) error
	Update(invoice *Invoice) error
	Delete(id string) error
}

type InvoiceQueryRepository interface {
	GetById(id string) (*Invoice, error)
	GetAll() ([]*Invoice, error)
}