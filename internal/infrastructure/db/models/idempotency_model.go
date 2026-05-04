package models

import (
	"time"

	"github.com/uptrace/bun"
)

type IdempotencyModel struct {
	bun.BaseModel `bun:"table:idempotency_records,alias:ir"`

	Key         string    `bun:"key,pk"`
	Fingerprint string    `bun:"fingerprint,notnull"`
	StatusCode  int       `bun:"status_code,notnull"`
	Body        []byte    `bun:"body"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
}
