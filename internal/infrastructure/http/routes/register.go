package routes

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/handlers"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	invoiceRoutes "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes/invoice"
	invoiceItemRoutes "github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/routes/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

// RegisterDomainRoutes registra rutas agrupadas por dominio de negocio.
func RegisterDomainRoutes(
	router fiber.Router,
	createH *handlers.CreateInvoiceHandler,
	updateH *handlers.UpdateInvoiceHandler,
	getAllH *handlers.GetAllInvoicesHandler,
	getByIDH *handlers.GetInvoiceByIDHandler,
	payH *handlers.PayInvoiceHandler,
	cancelH *handlers.CancelInvoiceHandler,
	createItemH *handlers.CreateInvoiceItemHandler,
	getAllItemH *handlers.GetAllInvoiceItemsHandler,
	getItemByIDH *handlers.GetInvoiceItemByIDHandler,
	updateItemH *handlers.UpdateInvoiceItemHandler,
	deleteItemH *handlers.DeleteInvoiceItemHandler,
	idempotencyRepo ports.IdempotencyRepository,
	catalogH *handlers.CatalogHandler,
	auth fiber.Handler,
) {
	invoiceRoutes.Register(router, createH, updateH, getAllH, getByIDH, payH, cancelH, idempotencyRepo, auth)
	invoiceItemRoutes.Register(router, createItemH, getAllItemH, getItemByIDH, updateItemH, deleteItemH, idempotencyRepo, auth)

	// Catálogos mínimos del tenant: el usuario elige clientes/productos legibles.
	customers := router.Group("/customers", auth)
	customers.Get("/", catalogH.ListCustomers)
	customers.Post("/", middleware.Idempotency(idempotencyRepo), catalogH.CreateCustomer)
	products := router.Group("/products", auth)
	products.Get("/", catalogH.ListProducts)
	products.Post("/", middleware.Idempotency(idempotencyRepo), catalogH.CreateProduct)

	// Metrics endpoint for observability
	metricsHandler := handlers.NewMetricsHandler()
	router.Get("/metrics", metricsHandler.Handle)
}
