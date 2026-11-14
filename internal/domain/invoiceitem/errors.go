package invoiceitem

import (
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/domainerr"
)

var ErrNotFound = errors.New("invoice item not found")

// Errores de validación (HTTP 400). El texto es contrato con el frontend: no cambiarlo.
var (
	ErrInvoiceIDRequired = domainerr.NewValidation("invoice_id is required")
	ErrItemIDRequired    = domainerr.NewValidation("item_id is required")
	ErrQuantityInvalid   = domainerr.NewValidation("quantity must be greater than 0")
	ErrUnitPriceInvalid  = domainerr.NewValidation("unit_price must be greater than 0")
	ErrTaxCodeRequired   = domainerr.NewValidation("tax code is required")
	ErrTaxRateInvalid    = domainerr.NewValidation("tax rate must be greater than or equal to 0")
	ErrTaxKindInvalid    = domainerr.NewValidation("tax kind must be DEBIT or CREDIT")
	// D4-02: un mismo código (sin distinguir mayúsculas ni espacios) no se repite en una línea.
	ErrTaxCodeDuplicated = domainerr.NewValidation("tax code is repeated in the same item")
)
