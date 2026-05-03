package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/db/models"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/uptrace/bun"
)

type idempotencyRepo struct {
	db *bun.DB
}

func NewIdempotencyRepo(db *bun.DB) ports.IdempotencyRepository {
	return &idempotencyRepo{db: db}
}

func (r *idempotencyRepo) FindByKey(key string) (*ports.IdempotencyRecord, error) {
	m := new(models.IdempotencyModel)
	ttlThreshold := time.Now().Add(-24 * time.Hour)

	err := r.db.NewSelect().
		Model(m).
		Where("key = ? AND created_at > ?", key, ttlThreshold).
		Scan(context.Background())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ports.IdempotencyRecord{
		Key:         m.Key,
		Fingerprint: m.Fingerprint,
		StatusCode:  m.StatusCode,
		Body:        m.Body,
		CreatedAt:   m.CreatedAt,
	}, nil
}

func (r *idempotencyRepo) Claim(record *ports.IdempotencyRecord) (bool, error) {
	m := &models.IdempotencyModel{
		Key:         record.Key,
		Fingerprint: record.Fingerprint,
		StatusCode:  record.StatusCode,
		Body:        record.Body,
		CreatedAt:   record.CreatedAt,
	}

	res, err := r.db.NewInsert().
		Model(m).
		On("CONFLICT (key) DO NOTHING").
		Exec(context.Background())
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *idempotencyRepo) Finalize(key string, statusCode int, body []byte) error {
	_, err := r.db.NewUpdate().
		Model((*models.IdempotencyModel)(nil)).
		Set("status_code = ?", statusCode).
		Set("body = ?", body).
		Where("key = ?", key).
		Exec(context.Background())
	return err
}
