package invoice

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// Resource es el recurso de permisos del dominio factura.
const Resource = "invoices"

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
	auth fiber.Handler,
) {
	// Toda operación de facturas exige un token válido; el tenant sale del token.
	// Cada operación exige además su permiso (invoices/read o invoices/manage),
	// antes de la idempotencia para que una petición denegada no reserve la clave.
	read := middleware.RequirePermission(Resource, middleware.ActionRead)
	manage := middleware.RequirePermission(Resource, middleware.ActionManage)
	g := router.Group("/invoices", auth)
	g.Post("/", manage, middleware.Idempotency(idempotencyRepo), createH.Handle)
	g.Put("/:id", manage, middleware.Idempotency(idempotencyRepo), updateH.Handle)
	g.Get("/", read, getAllH.Handle)
	g.Get("/:id", read, getByIDH.Handle)
	g.Patch("/:id/pay", manage, middleware.Idempotency(idempotencyRepo), payH.Handle)
	g.Patch("/:id/cancel", manage, middleware.Idempotency(idempotencyRepo), cancelH.Handle)
}
