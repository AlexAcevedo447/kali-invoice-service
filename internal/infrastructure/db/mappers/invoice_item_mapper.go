package mappers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
)

func InvoiceItemToModel(item *invoiceitem.InvoiceItem) *models.InvoiceItemModel {
	taxes := make([]models.InvoiceItemTaxModel, 0, len(item.Taxes))
	for _, t := range item.Taxes {
		invoiceID := t.InvoiceID
		invoiceItemID := t.InvoiceItemID
		taxes = append(taxes, models.InvoiceItemTaxModel{
			ID:            t.ID,
			InvoiceID:     &invoiceID,
			InvoiceItemID: &invoiceItemID,
			Code:          t.Code,
			Kind:          string(t.Kind),
			Rate:          t.Rate,
			Amount:        t.Amount,
		})
	}

	return &models.InvoiceItemModel{
		ID:        item.ID,
		InvoiceID: item.InvoiceID,
		ItemID:    item.ItemID,
		Quantity:  item.Quantity,
		UnitPrice: item.UnitPrice,
		TaxTotal:  item.TaxTotal,
		Subtotal:  item.Subtotal,
		Total:     item.Total,
		Taxes:     taxes,
	}
}

func ModelToInvoiceItem(m *models.InvoiceItemModel) *invoiceitem.InvoiceItem {
	taxes := make([]invoiceitem.Tax, 0, len(m.Taxes))
	for _, t := range m.Taxes {
		invoiceID := ""
		if t.InvoiceID != nil {
			invoiceID = *t.InvoiceID
		}
		invoiceItemID := ""
		if t.InvoiceItemID != nil {
			invoiceItemID = *t.InvoiceItemID
		}
		taxes = append(taxes, invoiceitem.Tax{
			ID:            t.ID,
			InvoiceID:     invoiceID,
			InvoiceItemID: invoiceItemID,
			Code:          t.Code,
			Kind:          invoiceitem.TaxKind(t.Kind),
			Rate:          t.Rate,
			Amount:        t.Amount,
		})
	}

	return &invoiceitem.InvoiceItem{
		ID:        m.ID,
		InvoiceID: m.InvoiceID,
		ItemID:    m.ItemID,
		Quantity:  m.Quantity,
		UnitPrice: m.UnitPrice,
		Taxes:     taxes,
		TaxTotal:  m.TaxTotal,
		Subtotal:  m.Subtotal,
		Total:     m.Total,
	}
}
