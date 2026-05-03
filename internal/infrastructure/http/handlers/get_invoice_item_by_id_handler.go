package handlers

import (
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/query"
	"github.com/gofiber/fiber/v2"
)

type GetInvoiceItemByIDHandler struct {
	getByIDQuery *appquery.GetInvoiceItemByIDQuery
}

func NewGetInvoiceItemByIDHandler(getByIDQuery *appquery.GetInvoiceItemByIDQuery) *GetInvoiceItemByIDHandler {
	return &GetInvoiceItemByIDHandler{getByIDQuery: getByIDQuery}
}

func (h *GetInvoiceItemByIDHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}
	item, err := h.getByIDQuery.Execute(id)
	if err != nil {
		return mapDomainError(err)
	}
	return c.JSON(item)
}
