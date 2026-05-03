package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type PayInvoiceHandler struct {
	payCmd       *appcommand.PayInvoiceCommand
	getByIDQuery *appquery.GetInvoiceByIDQuery
}

func NewPayInvoiceHandler(
	payCmd *appcommand.PayInvoiceCommand,
	getByIDQuery *appquery.GetInvoiceByIDQuery,
) *PayInvoiceHandler {
	return &PayInvoiceHandler{payCmd: payCmd, getByIDQuery: getByIDQuery}
}

func (h *PayInvoiceHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}
	if err := h.payCmd.Execute(id); err != nil {
		logger.Error("failed to pay invoice", logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoicesPaid()
	logger.Info("invoice paid successfully", logger.Fields{
		"invoice_id": id,
	})

	inv, err := h.getByIDQuery.Execute(id)
	if err != nil {
		logger.Error("failed to retrieve paid invoice", logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}
	return c.JSON(inv)
}