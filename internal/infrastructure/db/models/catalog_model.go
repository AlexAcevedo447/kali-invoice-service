package models

import (
	"time"

	"github.com/uptrace/bun"
)

type CustomerModel struct {
	bun.BaseModel `bun:"table:customers,alias:cu"`

	ID        string    `bun:"id,pk,type:uuid"`
	TenantID  string    `bun:"tenant_id,notnull,type:text"`
	Name      string    `bun:"name,notnull"`
	Document  string    `bun:"document,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull"`
}

type ProductModel struct {
	bun.BaseModel `bun:"table:products,alias:pr"`

	ID        string    `bun:"id,pk,type:uuid"`
	TenantID  string    `bun:"tenant_id,notnull,type:text"`
	Code      string    `bun:"code,notnull"`
	Name      string    `bun:"name,notnull"`
	UnitPrice float64   `bun:"unit_price,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull"`
}
