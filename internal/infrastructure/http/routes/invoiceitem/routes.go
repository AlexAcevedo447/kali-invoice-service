package invoiceitem

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Register agrupa los endpoints CRUD del dominio invoiceitem.
func Register(
	router fiber.Router,
	createH *handlers.CreateInvoiceItemHandler,
	getAllH *handlers.GetAllInvoiceItemsHandler,
	getByIDH *handlers.GetInvoiceItemByIDHandler,
	updateH *handlers.UpdateInvoiceItemHandler,
	deleteH *handlers.DeleteInvoiceItemHandler,
	idempotencyRepo ports.IdempotencyRepository,
) {
	g := router.Group("/invoice-items")
	g.Post("/", middleware.Idempotency(idempotencyRepo), createH.Handle)
	g.Get("/", getAllH.Handle)
	g.Get("/:id", getByIDH.Handle)
	g.Put("/:id", middleware.Idempotency(idempotencyRepo), updateH.Handle)
	g.Delete("/:id", middleware.Idempotency(idempotencyRepo), deleteH.Handle)
}