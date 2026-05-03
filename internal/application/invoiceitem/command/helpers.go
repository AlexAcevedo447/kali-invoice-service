package command

import "github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoiceitem"

func mapTaxInputs(inputs []InvoiceItemTaxInput) []invoiceitem.TaxInput {
	taxes := make([]invoiceitem.TaxInput, 0, len(inputs))
	for _, t := range inputs {
		taxes = append(taxes, invoiceitem.TaxInput{Code: t.Code, Kind: t.Kind, Rate: t.Rate})
	}
	return taxes
}
