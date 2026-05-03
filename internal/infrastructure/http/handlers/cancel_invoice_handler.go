package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	appquery "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/query"
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
	return handleStatusTransition(
		c,
		h.cancelCmd.Execute,
		h.getByIDQuery.Execute,
		metrics.IncrementInvoicesCanceled,
		"failed to cancel invoice",
		"invoice canceled successfully",
		"failed to retrieve canceled invoice",
	)
}
