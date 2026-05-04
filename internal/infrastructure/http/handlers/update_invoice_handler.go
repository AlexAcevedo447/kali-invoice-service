package handlers

import (
	"time"

	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	"github.com/gofiber/fiber/v2"
)

type UpdateInvoiceHandler struct {
	updateCmd *appcommand.UpdateInvoiceCommand
}

func NewUpdateInvoiceHandler(updateCmd *appcommand.UpdateInvoiceCommand) *UpdateInvoiceHandler {
	return &UpdateInvoiceHandler{updateCmd: updateCmd}
}

type updateInvoiceRequest struct {
	CustomerID *string    `json:"customer_id,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
}

func (h *UpdateInvoiceHandler) Handle(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}

	var req updateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.CustomerID != nil && *req.CustomerID != "" {
		if err := validateUUID(*req.CustomerID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "customer_id must be a valid UUID")
		}
	}

	inv, err := h.updateCmd.Execute(id, appcommand.UpdateInvoiceInput{
		CustomerID: req.CustomerID,
		DueDate:    req.DueDate,
	})
	if err != nil {
		return mapDomainError(err)
	}
	return c.JSON(inv)
}
