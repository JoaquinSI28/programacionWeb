package logic

import "errors"

var ErrValidacion = errors.New("datos inválidos")

type ErrorValidacion struct {
	Mensaje string
}

func (e *ErrorValidacion) Error() string {
	return e.Mensaje
}

func (e *ErrorValidacion) Unwrap() error {
	return ErrValidacion
}
