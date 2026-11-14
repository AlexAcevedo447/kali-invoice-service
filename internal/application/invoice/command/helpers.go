package command

import (
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/catalog"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/ports"
)

// ensureCustomerInTenant rechaza clientes inexistentes o de otra empresa.
func ensureCustomerInTenant(customers ports.CustomerRepository, tenantID, customerID string) error {
	ok, err := customers.Exists(tenantID, customerID)
	if err != nil {
		return err
	}
	if !ok {
		return catalog.ErrCustomerNotFound
	}
	return nil
}

// ensureProductsInTenant rechaza productos inexistentes o de otra empresa.
func ensureProductsInTenant(products ports.ProductRepository, tenantID string, ids []string) error {
	missing, err := products.MissingIDs(tenantID, ids)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return catalog.ErrProductNotFound
	}
	return nil
}

func mapInvoiceItemTaxes(inputs []InvoiceItemTaxInput) []invoiceitem.TaxInput {
	taxes := make([]invoiceitem.TaxInput, 0, len(inputs))
	for _, tax := range inputs {
		taxes = append(taxes, invoiceitem.TaxInput{Code: tax.Code, Kind: tax.Kind, Rate: tax.Rate})
	}
	return taxes
}
