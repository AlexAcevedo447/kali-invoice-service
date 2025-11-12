package http

import (
	"encoding/json"
	"net/http"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/command"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/dto"
	"github.com/AlexAcevedo447/kali-invoice-service/internal/application/query"
	"github.com/go-chi/chi/v5"
)

type InvoiceHandler struct {
	CreateCmd *command.AppCreateInvoiceCommand
	GetByIdQuery *query.AppGetByIdInvoiceQuery
}

func NewInvoiceHandler(
	createCmd *command.AppCreateInvoiceCommand, 
	getByIdQuery *query.AppGetByIdInvoiceQuery,
) *InvoiceHandler{
	return &InvoiceHandler{
		CreateCmd: createCmd,
		GetByIdQuery: getByIdQuery,
	}
}

func (handler * InvoiceHandler) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request){
	var dto dto.AppCreateInvoiceDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}

	invoice, err := handler.CreateCmd.Execute(dto);
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

func (handler *InvoiceHandler) GetInvoiceByIdHandler(w http.ResponseWriter, r *http.Request) {
	id := dto.AppGetByIdInvoiceDTO{
		Id: r.URL.Query().Get("id"),
	}

    invoice, err := handler.GetByIdQuery.Execute(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(invoice)
}

func (handler *InvoiceHandler) Router() RouteConfig {
	router := chi.NewRouter()

	router.Post("/", handler.CreateInvoiceHandler)
	router.Get("/{id}", handler.GetInvoiceByIdHandler)

	return RouteConfig{
		BasePath: "/invoices",
		Handler: router,
	}
}