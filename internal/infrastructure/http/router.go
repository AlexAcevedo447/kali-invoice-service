package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RouteConfig struct {
	BasePath string
	Handler  http.Handler
}

type RouterHandler interface {
	Routes() RouteConfig
}

func NewKaliInvoiceRouter(handlers ...RouterHandler) http.Handler {
	r := chi.NewRouter()

	for _, h := range handlers {
		cfg := h.Routes()
		r.Mount(cfg.BasePath, cfg.Handler)
	}

	return r
}
