package logic

import "errors"

var ErrValidacion = errors.New("datos inválidos")
var ErrDuplicado = errors.New("el nombre de usuario o el email ya está registrado")
var ErrReferencia = errors.New("el registro referenciado no existe o tiene datos relacionados")

type ErrorValidacion struct {
	Mensaje string
}

func (e *ErrorValidacion) Error() string {
	return e.Mensaje
}

func (e *ErrorValidacion) Unwrap() error {
	return ErrValidacion
}
