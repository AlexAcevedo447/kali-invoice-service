package invoiceitem

type TaxInput struct {
	Code string
	Kind string
	Rate float64
}

type TaxKind string

const (
	TaxKindDebit  TaxKind = "DEBIT"
	TaxKindCredit TaxKind = "CREDIT"
)

type Tax struct {
	ID            string
	InvoiceID     string
	InvoiceItemID string
	Code          string
	Kind          TaxKind
	Rate          float64
	Amount        float64
}
