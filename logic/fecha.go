package logic

import (
	"time"
	_ "time/tzdata"
)

var zonaBiblioteca = func() *time.Location {
	zona, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		panic(err)
	}
	return zona
}()

func Hoy() string { return time.Now().In(zonaBiblioteca).Format(time.DateOnly) }

func validarFecha(fecha Date, campo string) error {
	if fecha.IsZero() || fecha.Year() < 1 || fecha.Year() > 9999 {
		return &ErrorValidacion{Mensaje: campo + " debe ser una fecha válida"}
	}
	if fecha.Format(time.DateOnly) > Hoy() {
		return &ErrorValidacion{Mensaje: campo + " no puede ser futura"}
	}
	return nil
}
