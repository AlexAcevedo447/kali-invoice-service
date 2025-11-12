package entities

import (
	"errors"

	"github.com/google/uuid"
)

type Item struct {
	Id string  `gorm:"primaryKey;type:uuid" json:"id"`
	Name string  `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	UnitPrice float64 `gorm:"not null" json:"unit_price"`
	TaxRate float64 `gorm:"not null" json:"tax_rate"` // e.g. 0.19 = 19%
}

func NewItem(dto *NewItemDTO)(*Item, error){
	if dto.Name == "" {
		return nil, errors.New("name required")
	}

	if dto.UnitPrice <= 0 {
		return nil, errors.New("price must be positive")
	}

	item := &Item{
		Id: uuid.New().String(),
		Name: dto.Name,
		Description: dto.Description,
		UnitPrice: dto.UnitPrice,
		TaxRate: dto.TaxRate,
	}

	return item, nil
}

func (i *Item) PriceWithTax() float64 {
	return i.UnitPrice * (1 + i.TaxRate)
}