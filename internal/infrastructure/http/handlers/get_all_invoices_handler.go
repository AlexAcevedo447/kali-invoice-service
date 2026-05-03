package handlers

import (
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

	invoices, err := h.getAllQuery.Execute(p)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(invoices)
}