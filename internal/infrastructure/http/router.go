package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type IRouteConfig struct {
	BasePath string
	Handler  http.Handler
}

type RouterHandler interface {
	Router() IRouteConfig
}

func NewKaliInvoiceRouter(handlers ...RouterHandler) http.Handler {
	r := chi.NewRouter()

	for _, h := range handlers {
		cfg := h.Router()
		r.Mount(cfg.BasePath, cfg.Handler)
	}

	return r
}
