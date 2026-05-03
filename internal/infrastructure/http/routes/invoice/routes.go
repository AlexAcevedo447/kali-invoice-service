package invoice

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Register agrupa los endpoints del dominio factura.
func Register(
	router fiber.Router,
	createH *handlers.CreateInvoiceHandler,
	updateH *handlers.UpdateInvoiceHandler,
	getAllH *handlers.GetAllInvoicesHandler,
	getByIDH *handlers.GetInvoiceByIDHandler,
	payH *handlers.PayInvoiceHandler,
	cancelH *handlers.CancelInvoiceHandler,
	idempotencyRepo ports.IdempotencyRepository,
) {
	g := router.Group("/invoices")
	g.Post("/", middleware.Idempotency(idempotencyRepo), createH.Handle)
	g.Put("/:id", middleware.Idempotency(idempotencyRepo), updateH.Handle)
	g.Get("/", getAllH.Handle)
	g.Get("/:id", getByIDH.Handle)
	g.Patch("/:id/pay", middleware.Idempotency(idempotencyRepo), payH.Handle)
	g.Patch("/:id/cancel", middleware.Idempotency(idempotencyRepo), cancelH.Handle)
}
