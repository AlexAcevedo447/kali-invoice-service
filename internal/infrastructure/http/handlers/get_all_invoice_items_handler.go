package handlers

import (
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
	"github.com/gofiber/fiber/v2"
)

type GetAllInvoiceItemsHandler struct {
	getAllQuery *appquery.GetAllInvoiceItemsQuery
}

func NewGetAllInvoiceItemsHandler(getAllQuery *appquery.GetAllInvoiceItemsQuery) *GetAllInvoiceItemsHandler {
	return &GetAllInvoiceItemsHandler{getAllQuery: getAllQuery}
}

func (h *GetAllInvoiceItemsHandler) Handle(c *fiber.Ctx) error {
	p := ports.Pagination{Page: c.QueryInt("page", 1), PageSize: c.QueryInt("page_size", 20)}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	items, err := h.getAllQuery.Execute(middleware.TenantID(c), p)
	if err != nil {
		return mapDomainError(err)
	}
	return c.JSON(items)
}
