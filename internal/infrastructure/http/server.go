package http

import (
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: centralErrorHandler,
	})
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	return app
}

// internalErrorMessage es el único texto que el cliente ve en una respuesta 500.
const internalErrorMessage = "internal server error"

// centralErrorHandler responde {"error": msg}. Los 4xx y los 5xx distintos de
// 500 (p. ej. 503 del JWKS) conservan su mensaje, que es útil para el cliente.
// Un 500 siempre responde un mensaje genérico; el detalle queda en el log.
func centralErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := internalErrorMessage

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		msg = e.Message
	}

	if code == fiber.StatusInternalServerError {
		logger.Error("internal server error", logger.Fields{
			"method": c.Method(),
			"path":   c.Path(),
			"error":  err.Error(),
		})
		msg = internalErrorMessage
	}

	return c.Status(code).JSON(fiber.Map{"error": msg})
}
