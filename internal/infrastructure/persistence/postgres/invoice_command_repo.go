package postgres

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/contracts"
	"gorm.io/gorm"
)

type InvoiceCommandRepository struct {
	db *gorm.DB
}

func NewInvoiceCommandRepository(db *gorm.DB) *InvoiceCommandRepository {
	return &InvoiceCommandRepository{db: db}
}

func (r *InvoiceCommandRepository) Save(invoice *entities.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *InvoiceCommandRepository) Update(invoice *entities.Invoice) error {
	return r.db.Save(invoice).Error
}

// Garantiza que implementa las interfaces del dominio
var (
	_ contracts.ICreateInvoiceCommandRepository   = (*InvoiceCommandRepository)(nil)
)
