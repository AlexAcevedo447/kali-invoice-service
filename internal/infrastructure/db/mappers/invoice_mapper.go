package mappers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
)

func InvoiceToModel(inv *invoice.Invoice) *models.InvoiceModel {
	items := make([]models.InvoiceItemModel, 0, len(inv.Items))
	for _, it := range inv.Items {
		taxes := make([]models.InvoiceItemTaxModel, 0, len(it.Taxes))
		for _, tax := range it.Taxes {
			invoiceID := tax.InvoiceID
			invoiceItemID := tax.InvoiceItemID
			taxes = append(taxes, models.InvoiceItemTaxModel{
				ID:            tax.ID,
				InvoiceID:     &invoiceID,
				InvoiceItemID: &invoiceItemID,
				Code:          tax.Code,
				Kind:          string(tax.Kind),
				Rate:          tax.Rate,
				Amount:        tax.Amount,
			})
		}
		items = append(items, models.InvoiceItemModel{
			ID:        it.ID,
			InvoiceID: it.InvoiceID,
			ItemID:    it.ItemID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			TaxTotal:  it.TaxTotal,
			Subtotal:  it.Subtotal,
			Total:     it.Total,
			Taxes:     taxes,
		})
	}
	return &models.InvoiceModel{
		ID:         inv.ID,
		CustomerID: inv.CustomerID,
		IssueDate:  inv.IssueDate,
		DueDate:    inv.DueDate,
		Subtotal:   inv.Subtotal,
		TaxTotal:   inv.TaxTotal,
		Total:      inv.Total,
		Status:     string(inv.Status),
		CreatedAt:  inv.CreatedAt,
		UpdatedAt:  inv.UpdatedAt,
		Items:      items,
	}
}

func ModelToInvoice(m *models.InvoiceModel) *invoice.Invoice {
	items := make([]invoiceitem.InvoiceItem, 0, len(m.Items))
	for _, it := range m.Items {
		taxes := make([]invoiceitem.Tax, 0, len(it.Taxes))
		for _, tax := range it.Taxes {
			invoiceID := ""
			if tax.InvoiceID != nil {
				invoiceID = *tax.InvoiceID
			}
			invoiceItemID := ""
			if tax.InvoiceItemID != nil {
				invoiceItemID = *tax.InvoiceItemID
			}
			taxes = append(taxes, invoiceitem.Tax{
				ID:            tax.ID,
				InvoiceID:     invoiceID,
				InvoiceItemID: invoiceItemID,
				Code:          tax.Code,
				Kind:          invoiceitem.TaxKind(tax.Kind),
				Rate:          tax.Rate,
				Amount:        tax.Amount,
			})
		}
		items = append(items, invoiceitem.InvoiceItem{
			ID:        it.ID,
			InvoiceID: it.InvoiceID,
			ItemID:    it.ItemID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Taxes:     taxes,
			TaxTotal:  it.TaxTotal,
			Subtotal:  it.Subtotal,
			Total:     it.Total,
		})
	}
	return &invoice.Invoice{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		IssueDate:  m.IssueDate,
		DueDate:    m.DueDate,
		Subtotal:   m.Subtotal,
		TaxTotal:   m.TaxTotal,
		Total:      m.Total,
		Status:     invoice.Status(m.Status),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		Items:      items,
	}
}
