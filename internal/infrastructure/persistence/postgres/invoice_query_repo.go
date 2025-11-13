package postgres

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"gorm.io/gorm"
)

type InvoiceQueryRepository struct {
	db *gorm.DB
}

func NewInvoiceQueryRepository(db *gorm.DB) *InvoiceQueryRepository {
	return &InvoiceQueryRepository{db: db}
}

func (r *InvoiceQueryRepository) GetById(id string) (*entities.Invoice, error) {
	var inv entities.Invoice
	if err := r.db.Preload("Items").First(&inv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceQueryRepository) GetAll() ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// Garantiza que implementa las interfaces del dominio
var (
	_ domain.IGetAllInvoiceQueryRepository   = (*InvoiceQueryRepository)(nil)
	_ domain.IGetByIdInvoiceQueryRepository   = (*InvoiceQueryRepository)(nil)
)
