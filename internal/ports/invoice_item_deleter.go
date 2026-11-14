package ports

type InvoiceItemDeleter interface {
	DeleteByID(tenantID, id string) error
}
