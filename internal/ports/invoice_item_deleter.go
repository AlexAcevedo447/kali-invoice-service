package ports

type InvoiceItemDeleter interface {
	DeleteByID(id string) error
}
