package main

import (
	"errors"

	"bookfly.com/logic"
)

func mensajeValidacion(err error) string {
	var validacion *logic.ErrorValidacion
	if errors.As(err, &validacion) {
		return validacion.Mensaje
	}
	return "Datos inválidos"
}
