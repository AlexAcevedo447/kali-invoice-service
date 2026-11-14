package repositories

import "github.com/uptrace/bun"

// orderItemTaxes devuelve los impuestos de cada línea en el orden en que se
// registraron (CA-13.2). PUT /invoice-items/:id reemplaza la lista borrando e
// insertando, así que sin este orden Postgres podría devolverlos en otro.
func orderItemTaxes(q *bun.SelectQuery) *bun.SelectQuery {
	return q.OrderExpr("position ASC, id ASC")
}
