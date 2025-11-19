package commands

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/contracts"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/models"
)

type InvoiceFactoryCommandMapper struct {
}

func (command *InvoiceFactoryCommandMapper) InvoiceToModel(invoice *entities.Invoice) models.InvoiceModel {
	items := make([]models.InvoiceItemModel, 0)

	for _, item := range invoice.Items {
		items = append(items, models.InvoiceItemModel{
			Id: item.Id,
			InvoiceId: item.InvoiceId,
			ItemId: item.ItemId,
			Quantity: item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal: item.Subtotal,
		})
	}

	return models.InvoiceModel{
		Id: invoice.Id,
		CustomerId: invoice.CustomerId,
		IssueDate: invoice.IssueDate,
		DueDate: invoice.DueDate,
		Items: items,
		Total: invoice.Total,
		Status: string(invoice.Status),
	}
}

func (command *InvoiceFactoryCommandMapper) ModelToInvoice(model models.InvoiceModel) *entities.Invoice{
	items := make([]entities.InvoiceItem, 0)

	for _, item := range model.Items {
		items = append(items, entities.InvoiceItem{
			Id: item.Id,
			InvoiceId: item.InvoiceId,
			ItemId: item.ItemId,
			Quantity: item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal: item.Subtotal,
		})
	}

	return &entities.Invoice{
		Id: model.Id,
		CustomerId: model.CustomerId,
		IssueDate: model.IssueDate,
		DueDate: model.DueDate,
		Items: items,
		Total: model.Total,
		Status: entities.InvoiceStatus(model.Status),
	}
}

var _ contracts.IInvoiceFactoryCommandMapper = (*InvoiceFactoryCommandMapper)(nil)