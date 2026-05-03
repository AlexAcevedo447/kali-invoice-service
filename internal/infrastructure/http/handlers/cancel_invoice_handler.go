package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type CancelInvoiceHandler struct {
	cancelCmd    *appcommand.CancelInvoiceCommand
	getByIDQuery *appquery.GetInvoiceByIDQuery
}

func NewCancelInvoiceHandler(
	cancelCmd *appcommand.CancelInvoiceCommand,
	getByIDQuery *appquery.GetInvoiceByIDQuery,
) *CancelInvoiceHandler {
	return &CancelInvoiceHandler{cancelCmd: cancelCmd, getByIDQuery: getByIDQuery}
}

func (h *CancelInvoiceHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}
	if err := h.cancelCmd.Execute(id); err != nil {
		logger.Error("failed to cancel invoice", logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoicesCanceled()
	logger.Info("invoice canceled successfully", logger.Fields{
		"invoice_id": id,
	})

	inv, err := h.getByIDQuery.Execute(id)
	if err != nil {
		logger.Error("failed to retrieve canceled invoice", logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}
	return c.JSON(inv)
}