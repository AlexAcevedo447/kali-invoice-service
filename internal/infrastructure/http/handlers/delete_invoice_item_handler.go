package handlers

import (
	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoiceitem/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type DeleteInvoiceItemHandler struct {
	deleteCmd *appcommand.DeleteInvoiceItemCommand
}

func NewDeleteInvoiceItemHandler(deleteCmd *appcommand.DeleteInvoiceItemCommand) *DeleteInvoiceItemHandler {
	return &DeleteInvoiceItemHandler{deleteCmd: deleteCmd}
}

func (h *DeleteInvoiceItemHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}
	if err := h.deleteCmd.Execute(id); err != nil {
		logger.Error("failed to delete invoice item", logger.Fields{
			"item_id": id,
			"error":   err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoiceItemsDeleted()
	logger.Info("invoice item deleted successfully", logger.Fields{
		"item_id": id,
	})
	return c.SendStatus(fiber.StatusNoContent)
}