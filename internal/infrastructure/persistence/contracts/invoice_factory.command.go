package contracts

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/persistence/models"
)

type IInvoiceFactoryCommandMapper interface {
	InvoiceToModel(invoice *entities.Invoice) models.InvoiceModel
	ModelToInvoice(model models.InvoiceModel) *entities.Invoice
}