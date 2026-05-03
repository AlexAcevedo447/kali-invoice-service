package command

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

func mapInvoiceItemTaxes(inputs []InvoiceItemTaxInput) []invoiceitem.TaxInput {
	taxes := make([]invoiceitem.TaxInput, 0, len(inputs))
	for _, tax := range inputs {
		taxes = append(taxes, invoiceitem.TaxInput{Code: tax.Code, Kind: tax.Kind, Rate: tax.Rate})
	}
	return taxes
}
