// Package domainerr define tipos de error compartidos por los paquetes de dominio.
package domainerr

import "errors"

// ValidationError representa un dato de entrada inválido según las reglas del
// dominio. La capa HTTP lo traduce a 400 conservando el mensaje exacto, que el
// frontend usa como clave de traducción.
type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string { return e.msg }

// NewValidation crea un error de validación. Se usa para declarar sentinelas
// (var ErrX = domainerr.NewValidation("...")) comparables con errors.Is.
func NewValidation(msg string) error {
	return &ValidationError{msg: msg}
}

// AsValidation indica si err (o algún error envuelto) es de validación y lo devuelve.
func AsValidation(err error) (*ValidationError, bool) {
	var v *ValidationError
	if errors.As(err, &v) {
		return v, true
	}
	return nil, false
}
