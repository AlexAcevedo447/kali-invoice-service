package invoice

import "errors"

var (
	ErrNotFound        = errors.New("invoice not found")
	ErrAlreadyPaid     = errors.New("invoice is already paid")
	ErrAlreadyCanceled = errors.New("invoice is already canceled")
)
