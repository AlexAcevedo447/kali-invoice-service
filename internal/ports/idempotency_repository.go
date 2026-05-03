package ports

import "time"

// IdempotencyRecord almacena la respuesta cacheada de una operación idempotente.
type IdempotencyRecord struct {
	Key         string
	Fingerprint string
	StatusCode  int
	Body        []byte
	CreatedAt   time.Time
}

type IdempotencyRepository interface {
	// FindByKey busca un registro activo (dentro del TTL de 24h). Retorna nil si no existe.
	FindByKey(key string) (*IdempotencyRecord, error)
	// Claim intenta reservar una llave idempotente de forma atómica.
	// Retorna true si la reserva fue exitosa; false si la llave ya existía.
	Claim(record *IdempotencyRecord) (bool, error)
	// Finalize guarda la respuesta final de una operación ya reservada.
	Finalize(key string, statusCode int, body []byte) error
}
