package handlers

import (
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/domainerr"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/http/middleware"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// InternalErrorMessage es el único texto que el cliente ve en una respuesta 500.
const InternalErrorMessage = "internal server error"

func validateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id: must be a valid UUID")
	}
	return nil
}

// domainErrorStatus asocia cada error sentinela del dominio con su código HTTP.
// Se evalúa con errors.Is para que los errores envueltos (fmt.Errorf("...: %w", err))
// conserven su código. La respuesta usa el texto del sentinela, que es el que el
// frontend traduce por coincidencia exacta.
var domainErrorStatus = []struct {
	err    error
	status int
}{
	{invoice.ErrNotFound, fiber.StatusNotFound},
	{invoiceitem.ErrNotFound, fiber.StatusNotFound},
	{invoice.ErrAlreadyPaid, fiber.StatusUnprocessableEntity},
	{invoice.ErrAlreadyCanceled, fiber.StatusUnprocessableEntity},
	// Referencias a clientes/productos inexistentes o de otra empresa.
	{catalog.ErrCustomerNotFound, fiber.StatusUnprocessableEntity},
	{catalog.ErrProductNotFound, fiber.StatusUnprocessableEntity},
	{catalog.ErrCustomerDocDuplicated, fiber.StatusConflict},
	{catalog.ErrProductCodeDuplicated, fiber.StatusConflict},
	{invoice.ErrTenantRequired, fiber.StatusUnauthorized},
	{catalog.ErrTenantRequired, fiber.StatusUnauthorized},
}

func mapDomainError(err error) error {
	for _, m := range domainErrorStatus {
		if errors.Is(err, m.err) {
			return fiber.NewError(m.status, m.err.Error())
		}
	}
	// Cualquier error de validación del dominio (invoice, invoiceitem, catalog) es 400.
	if v, ok := domainerr.AsValidation(err); ok {
		return fiber.NewError(fiber.StatusBadRequest, v.Error())
	}
	// Solo lo realmente inesperado (BD, red, bugs) llega a 500. El detalle se
	// registra en el log; al cliente nunca se le exponen errores internos.
	logger.Error("unexpected error", logger.Fields{"error": err.Error()})
	return fiber.NewError(fiber.StatusInternalServerError, InternalErrorMessage)
}

func handleStatusTransition(
	c *fiber.Ctx,
	executeCmd func(tenantID, id string) error,
	getByIDQuery func(tenantID, id string) (*invoice.Invoice, error),
	incrementMetric func(),
	errorMsg string,
	successMsg string,
	retrieveErrorMsg string,
) error {
	id := c.Params("id")
	if err := validateUUID(id); err != nil {
		return err
	}

	tenantID := middleware.TenantID(c)
	if err := executeCmd(tenantID, id); err != nil {
		logger.Error(errorMsg, logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}

	incrementMetric()
	logger.Info(successMsg, logger.Fields{
		"invoice_id": id,
	})

	inv, err := getByIDQuery(tenantID, id)
	if err != nil {
		logger.Error(retrieveErrorMsg, logger.Fields{
			"invoice_id": id,
			"error":      err.Error(),
		})
		return mapDomainError(err)
	}
	return c.JSON(inv)
}
