package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceByIDRepo struct {
	db *bun.DB
}

func NewInvoiceByIDRepo(db *bun.DB) ports.InvoiceByIDReader {
	return &invoiceByIDRepo{db: db}
}

func (r *invoiceByIDRepo) GetByID(id string) (*invoice.Invoice, error) {
	m := new(models.InvoiceModel)
	err := r.db.NewSelect().
		Model(m).
		Relation("Items").
		Relation("Items.Taxes").
		Relation("Taxes").
		Where("i.id = ?", id).
		Scan(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invoice.ErrNotFound
		}
		return nil, err
	}
	return mappers.ModelToInvoice(m), nil
}
