package handlers

import (
	"time"

	appcommand "github.com/AlexAcevedo447/kali-invoice-service/internal/application/invoice/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

type CreateInvoiceHandler struct {
	createCmd *appcommand.CreateInvoiceCommand
}

func NewCreateInvoiceHandler(createCmd *appcommand.CreateInvoiceCommand) *CreateInvoiceHandler {
	return &CreateInvoiceHandler{createCmd: createCmd}
}

type createInvoiceRequest struct {
	CustomerID string                        `json:"customer_id"`
	IssueDate  *time.Time                    `json:"issue_date,omitempty"`
	DueDate    *time.Time                    `json:"due_date,omitempty"`
	Items      []appcommand.InvoiceItemInput `json:"items"`
}

func (h *CreateInvoiceHandler) Handle(c *fiber.Ctx) error {
	var req createInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.CustomerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "customer_id is required")
	}
	if err := validateUUID(req.CustomerID); err != nil {
		return err
	}
	if len(req.Items) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "items are required")
	}

	// Validate item_id UUIDs
	for i, item := range req.Items {
		if item.ItemID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "items["+string(rune(i))+"].item_id is required")
		}
		if err := validateUUID(item.ItemID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "items["+string(rune(i))+"].item_id must be a valid UUID")
		}
	}

	inv, err := h.createCmd.Execute(appcommand.CreateInvoiceInput{
		CustomerID: req.CustomerID,
		IssueDate:  req.IssueDate,
		DueDate:    req.DueDate,
		Items:      req.Items,
	})
	if err != nil {
		logger.Error("failed to create invoice", logger.Fields{
			"customer_id": req.CustomerID,
			"error":       err.Error(),
		})
		return mapDomainError(err)
	}

	metrics.IncrementInvoicesCreated()
	logger.Info("invoice created successfully", logger.Fields{
		"invoice_id":  inv.ID,
		"customer_id": req.CustomerID,
		"items_count": len(req.Items),
		"subtotal":    inv.Subtotal,
		"tax_total":   inv.TaxTotal,
		"total":       inv.Total,
	})
	return c.Status(fiber.StatusCreated).JSON(inv)
}
