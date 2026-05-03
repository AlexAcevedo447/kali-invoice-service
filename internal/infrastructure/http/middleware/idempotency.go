package middleware

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Idempotency garantiza que operaciones mutables se procesen una sola vez por llave.
// Requiere el header Idempotency-Key en cada solicitud.
//
// Flujo:
//   - Claim atómico de la llave en DB para evitar doble ejecución concurrente.
//   - Reintento con mismo payload: replay de respuesta cacheada.
//   - Reintento con payload distinto: 422.
//   - Reintento mientras la primera petición sigue en progreso: 409.
func Idempotency(repo ports.IdempotencyRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("Idempotency-Key")
		if key == "" {
			logger.Error("idempotency key missing", logger.Fields{
				"method": c.Method(),
				"path":   c.Path(),
			})
			return fiber.NewError(fiber.StatusBadRequest, "Idempotency-Key header is required")
		}

		fingerprint := computeFingerprint(c.Method(), c.Path(), c.Body())

		record, err := repo.FindByKey(key)
		if err != nil {
			logger.Error("idempotency repository check failed", logger.Fields{
				"idempotency_key": key,
				"error":           err.Error(),
			})
			return fiber.NewError(fiber.StatusInternalServerError, "idempotency check failed")
		}

		if record != nil {
			return handleExistingRecord(c, key, fingerprint, record)
		}

		claimed, err := repo.Claim(&ports.IdempotencyRecord{
			Key:         key,
			Fingerprint: fingerprint,
			StatusCode:  0,
			Body:        nil,
			CreatedAt:   time.Now(),
		})
		if err != nil {
			logger.Error("idempotency claim failed", logger.Fields{
				"idempotency_key": key,
				"error":           err.Error(),
			})
			return fiber.NewError(fiber.StatusInternalServerError, "idempotency claim failed")
		}

		if !claimed {
			record, err = repo.FindByKey(key)
			if err != nil {
				logger.Error("idempotency reload failed", logger.Fields{
					"idempotency_key": key,
					"error":           err.Error(),
				})
				return fiber.NewError(fiber.StatusInternalServerError, "idempotency check failed")
			}
			if record != nil {
				return handleExistingRecord(c, key, fingerprint, record)
			}
			return fiber.NewError(fiber.StatusConflict, "idempotency request in progress")
		}

		metrics.IncrementIdempotencyMisses()
		if err := c.Next(); err != nil {
			return err
		}

		statusCode := c.Response().StatusCode()
		body := append([]byte(nil), c.Response().Body()...)
		if err := repo.Finalize(key, statusCode, body); err != nil {
			logger.Error("idempotency finalize failed", logger.Fields{
				"idempotency_key": key,
				"error":           err.Error(),
			})
		}

		logger.Info("idempotency record finalized", logger.Fields{
			"idempotency_key": key,
			"method":          c.Method(),
			"path":            c.Path(),
			"status_code":     statusCode,
		})

		return nil
	}
}

func handleExistingRecord(c *fiber.Ctx, key, fingerprint string, record *ports.IdempotencyRecord) error {
	if record.Fingerprint != fingerprint {
		logger.Warn("idempotency key reused with different payload", logger.Fields{
			"idempotency_key": key,
			"method":          c.Method(),
			"path":            c.Path(),
		})
		return fiber.NewError(fiber.StatusUnprocessableEntity,
			"idempotency key already used with a different payload")
	}

	if len(record.Body) == 0 {
		logger.Info("idempotency request still in progress", logger.Fields{
			"idempotency_key": key,
			"method":          c.Method(),
			"path":            c.Path(),
		})
		return fiber.NewError(fiber.StatusConflict, "idempotency request is still being processed")
	}

	metrics.IncrementIdempotencyHits()
	logger.Info("idempotency cache hit", logger.Fields{
		"idempotency_key": key,
		"method":          c.Method(),
		"path":            c.Path(),
		"status_code":     record.StatusCode,
	})
	c.Set("X-Idempotent-Replayed", "true")
	return c.Status(record.StatusCode).Send(record.Body)
}

func computeFingerprint(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(method + "|" + path + "|"))
	h.Write(body)
	return fmt.Sprintf("%x", h.Sum(nil))
}
