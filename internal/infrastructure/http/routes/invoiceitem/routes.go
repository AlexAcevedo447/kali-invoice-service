package invoiceitem

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Resource es el recurso de permisos del dominio ítem de factura.
const Resource = "invoice-items"

// Register agrupa los endpoints CRUD del dominio invoiceitem.
func Register(
	router fiber.Router,
	createH *handlers.CreateInvoiceItemHandler,
	getAllH *handlers.GetAllInvoiceItemsHandler,
	getByIDH *handlers.GetInvoiceItemByIDHandler,
	updateH *handlers.UpdateInvoiceItemHandler,
	deleteH *handlers.DeleteInvoiceItemHandler,
	idempotencyRepo ports.IdempotencyRepository,
	auth fiber.Handler,
) {
	read := middleware.RequirePermission(Resource, middleware.ActionRead)
	manage := middleware.RequirePermission(Resource, middleware.ActionManage)
	g := router.Group("/invoice-items", auth)
	g.Post("/", manage, middleware.Idempotency(idempotencyRepo), createH.Handle)
	g.Get("/", read, getAllH.Handle)
	g.Get("/:id", read, getByIDH.Handle)
	g.Put("/:id", manage, middleware.Idempotency(idempotencyRepo), updateH.Handle)
	g.Delete("/:id", manage, middleware.Idempotency(idempotencyRepo), deleteH.Handle)
}
