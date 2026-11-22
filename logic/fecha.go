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

// ValidarEdadAlLanzamiento comprueba la edad al día de publicación, usando fechas de calendario.
func ValidarEdadAlLanzamiento(nacimiento, lanzamiento Date) error {
	minima := nacimiento.AddDate(5, 0, 0)
	// Para nacimientos del 29 de febrero, el aniversario en años no bisiestos
	// se considera el último día de febrero.
	if nacimiento.Month() == time.February && nacimiento.Day() == 29 && minima.Month() == time.March {
		minima = minima.AddDate(0, 0, -1)
	}
	if lanzamiento.Format(time.DateOnly) < minima.Format(time.DateOnly) {
		return &ErrorValidacion{Mensaje: "El autor debe tener al menos 5 años cumplidos en la fecha de lanzamiento"}
	}
	return nil
}
