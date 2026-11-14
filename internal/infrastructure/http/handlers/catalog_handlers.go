package handlers

import (
	appcatalog "github.com/AlexAcevedo447/kali-invoice-service/internal/application/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

// CatalogHandler expone clientes y productos del tenant autenticado para que
// el usuario los elija por nombre, documento o código en vez de escribir UUIDs.
type CatalogHandler struct {
	service *appcatalog.Service
}

func NewCatalogHandler(service *appcatalog.Service) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func catalogQuery(c *fiber.Ctx) ports.CatalogQuery {
	return ports.CatalogQuery{
		Search:     c.Query("search"),
		Pagination: ports.Pagination{Page: c.QueryInt("page", 1), PageSize: c.QueryInt("page_size", 20)},
	}
}

type createCustomerRequest struct {
	Name     string `json:"name"`
	Document string `json:"document"`
}

func (h *CatalogHandler) ListCustomers(c *fiber.Ctx) error {
	items, total, err := h.service.ListCustomers(middleware.TenantID(c), catalogQuery(c))
	if err != nil {
		return mapDomainError(err)
	}
	c.Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(items)
}

func (h *CatalogHandler) CreateCustomer(c *fiber.Ctx) error {
	var req createCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	customer, err := h.service.CreateCustomer(middleware.TenantID(c), req.Name, req.Document)
	if err != nil {
		return mapDomainError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(customer)
}

type createProductRequest struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
}

func (h *CatalogHandler) ListProducts(c *fiber.Ctx) error {
	items, total, err := h.service.ListProducts(middleware.TenantID(c), catalogQuery(c))
	if err != nil {
		return mapDomainError(err)
	}
	c.Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(items)
}

func (h *CatalogHandler) CreateProduct(c *fiber.Ctx) error {
	var req createProductRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	product, err := h.service.CreateProduct(middleware.TenantID(c), req.Code, req.Name, req.UnitPrice)
	if err != nil {
		return mapDomainError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(product)
}
