package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemByIDRepo struct {
	db *bun.DB
}

func NewInvoiceItemByIDRepo(db *bun.DB) ports.InvoiceItemByIDReader {
	return &invoiceItemByIDRepo{db: db}
}

func (r *invoiceItemByIDRepo) GetByID(id string) (*invoiceitem.InvoiceItem, error) {
	m := new(models.InvoiceItemModel)
	err := r.db.NewSelect().
		Model(m).
		Relation("Taxes").
		Where("ii.id = ?", id).
		Scan(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, invoiceitem.ErrNotFound
		}
		return nil, err
	}
	return mappers.ModelToInvoiceItem(m), nil
}
