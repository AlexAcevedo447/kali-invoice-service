package invoice

import (
	"errors"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/domainerr"
)

var (
	ErrNotFound        = errors.New("invoice not found")
	ErrAlreadyPaid     = errors.New("invoice is already paid")
	ErrAlreadyCanceled = errors.New("invoice is already canceled")
	ErrTenantRequired  = errors.New("tenant is required")
)

// Errores de validación (HTTP 400). El texto es contrato con el frontend: no cambiarlo.
var (
	ErrCustomerIDRequired     = domainerr.NewValidation("customer_id is required")
	ErrItemsRequired          = domainerr.NewValidation("invoice must have at least one item")
	ErrDueDateBeforeIssueDate = domainerr.NewValidation("due_date cannot be before issue_date")
)
