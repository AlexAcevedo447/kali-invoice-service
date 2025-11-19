package models

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/entities"

type InvoiceItemModel struct {
	Id string `gorm:"primaryKey;type:uuid" json:"id"`
	InvoiceId string `gorm:"type:uuid;not null;index" json:"invoice_id"`
	ItemId string `gorm:"type:uuid;not null" json:"item_id"`
	Quantity float64 `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"not null" json:"unit_price"`
	Subtotal float64 `gorm:"not null" json:"subtotal"`
	Item *entities.Item `gorm:"foreignKey:ItemId" json:"item,omitempty"`
}