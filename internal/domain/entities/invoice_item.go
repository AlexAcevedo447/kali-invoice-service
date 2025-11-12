package entities

import (
	"errors"

	"github.com/google/uuid"
)

type InvoiceItem struct {
	Id string  `gorm:"primaryKey;type:uuid" json:"id"`
	InvoiceId string  `gorm:"type:uuid;not null;index" json:"invoice_id"`
	ItemId string  `gorm:"type:uuid;not null" json:"item_id"`
	Quantity int     `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"not null" json:"unit_price"`
	Subtotal float64 `gorm:"not null" json:"subtotal"`
	Item *Item   `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func NewInvoiceItem(dto NewInvoiceItemDTO) (*InvoiceItem, error){
	if dto.Item == nil {
		return nil, errors.New("item cannot be empty")
	}

	if dto.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	invoiceItem := &InvoiceItem{
		Id: uuid.New().String(),
		ItemId: dto.Item.Id,
		Quantity: dto.Quantity,
		UnitPrice: dto.UnitPrice,
	}

	invoiceItem.CalculateSubtotal()
	return invoiceItem, nil
}

func (ii *InvoiceItem) CalculateSubtotal(){
	ii.Subtotal = float64(ii.Quantity) * ii.UnitPrice
}