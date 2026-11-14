package handlers

import (
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/gofiber/fiber/v2"
)

type GetInvoiceByIDHandler struct {
	getByIDQuery *appquery.GetInvoiceByIDQuery
}

func NewGetInvoiceByIDHandler(getByIDQuery *appquery.GetInvoiceByIDQuery) *GetInvoiceByIDHandler {
	return &GetInvoiceByIDHandler{getByIDQuery: getByIDQuery}
}

func (h *GetInvoiceByIDHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}

	inv, err := h.getByIDQuery.Execute(middleware.TenantID(c), id)
	if err != nil {
		return mapDomainError(err)
	}
	return c.JSON(inv)
}
