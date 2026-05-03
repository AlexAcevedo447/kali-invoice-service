package repositories

import (
	"context"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/mappers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type invoiceItemListRepo struct {
	db *bun.DB
}

func NewInvoiceItemListRepo(db *bun.DB) ports.InvoiceItemListReader {
	return &invoiceItemListRepo{db: db}
}

func (r *invoiceItemListRepo) GetAll(p ports.Pagination) ([]*invoiceitem.InvoiceItem, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	offset := (p.Page - 1) * p.PageSize

	var ms []models.InvoiceItemModel
	err := r.db.NewSelect().
		Model(&ms).
		Relation("Taxes").
		Limit(p.PageSize).
		Offset(offset).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}

	items := make([]*invoiceitem.InvoiceItem, 0, len(ms))
	for i := range ms {
		items = append(items, mappers.ModelToInvoiceItem(&ms[i]))
	}
	return items, nil
}
