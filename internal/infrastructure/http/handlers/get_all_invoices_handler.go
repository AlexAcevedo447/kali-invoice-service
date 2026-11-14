package handlers

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"strconv"

	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

type GetAllInvoicesHandler struct {
	getAllQuery *appquery.GetAllInvoicesQuery
}

func NewGetAllInvoicesHandler(getAllQuery *appquery.GetAllInvoicesQuery) *GetAllInvoicesHandler {
	return &GetAllInvoicesHandler{getAllQuery: getAllQuery}
}

func (h *GetAllInvoicesHandler) Handle(c *fiber.Ctx) error {
	p := ports.Pagination{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}

	tenantID := middleware.TenantID(c)
	invoices, err := h.getAllQuery.Execute(tenantID, p)
	if err != nil {
		return mapDomainError(err)
	}
	total, err := h.getAllQuery.Total(tenantID)
	if err != nil {
		return mapDomainError(err)
	}
	// El cuerpo sigue siendo el arreglo de facturas; el total viaja en cabecera.
	c.Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(invoices)
}
