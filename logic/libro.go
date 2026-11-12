package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Date represents a calendar date.
type Date = time.Time

var ErrLibroNoEncontrado = errors.New("libro no encontrado")

type Libro struct {
	IDLibro          int32
	Titulo           string
	IDAutor          int32
	IDEditorial      int32
	FechaLanzamiento Date
	CantidadPaginas  int32
}

func validateTitulo(titulo string) error {
	if strings.TrimSpace(titulo) == "" {
		return &ErrorValidacion{Mensaje: "El título no puede estar vacío"}
	}
	return nil
}
func validateIDEditorial(idEditorial int32) error {
	if idEditorial <= 0 {
		return &ErrorValidacion{Mensaje: "El ID de la editorial debe ser un número positivo"}
	}
	return nil
}
func validateCantidadPaginas(cantidadPaginas int32) error {
	if cantidadPaginas <= 0 {
		return &ErrorValidacion{Mensaje: "La cantidad de páginas debe ser mayor a cero"}
	}
	return nil
}
func validateIDAutor(idAutor int32) error {
	if idAutor <= 0 {
		return &ErrorValidacion{Mensaje: "El ID del autor debe ser un número positivo"}
	}
	return nil
}

func CreateLibro(ctx context.Context, repo LibroRepository, autores AutorRepository, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error) {
	if err := validateTitulo(titulo); err != nil {
		return nil, fmt.Errorf("error al validar título: %w", err)
	}
	if err := validateIDAutor(idAutor); err != nil {
		return nil, fmt.Errorf("error al validar ID de autor: %w", err)
	}
	if err := validateIDEditorial(idEditorial); err != nil {
		return nil, fmt.Errorf("error al validar ID de editorial: %w", err)
	}
	if err := validarFecha(fechaLanzamiento, "La fecha de lanzamiento"); err != nil {
		return nil, fmt.Errorf("error al validar fecha de lanzamiento: %w", err)
	}
	if err := validateCantidadPaginas(cantidadPaginas); err != nil {
		return nil, fmt.Errorf("error al validar cantidad de páginas: %w", err)
	}

	if err := validarCronologiaLibro(ctx, autores, idAutor, fechaLanzamiento); err != nil {
		return nil, err
	}
	return repo.Create(ctx, idAutor, titulo, idEditorial, fechaLanzamiento, cantidadPaginas)
}

func UpdateLibro(ctx context.Context, repo LibroRepository, autores AutorRepository, id int32, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error) {
	if err := validateTitulo(titulo); err != nil {
		return nil, fmt.Errorf("error al validar título: %w", err)
	}
	if err := validateIDAutor(idAutor); err != nil {
		return nil, fmt.Errorf("error al validar ID de autor: %w", err)
	}
	if err := validateIDEditorial(idEditorial); err != nil {
		return nil, fmt.Errorf("error al validar ID de editorial: %w", err)
	}
	if err := validarFecha(fechaLanzamiento, "La fecha de lanzamiento"); err != nil {
		return nil, fmt.Errorf("error al validar fecha de lanzamiento: %w", err)
	}
	if err := validateCantidadPaginas(cantidadPaginas); err != nil {
		return nil, fmt.Errorf("error al validar cantidad de páginas: %w", err)
	}

	if _, err := GetLibro(ctx, repo, id); err != nil {
		return nil, err
	}

	if err := validarCronologiaLibro(ctx, autores, idAutor, fechaLanzamiento); err != nil {
		return nil, err
	}
	return repo.Update(ctx, id, idAutor, titulo, idEditorial, fechaLanzamiento, cantidadPaginas)
}

func GetLibro(ctx context.Context, repo LibroRepository, id int32) (*Libro, error) {
	libro, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener libro: %w", err)
	}
	return libro, nil
}

func GetLibros(ctx context.Context, repo LibroRepository) ([]*Libro, error) {
	libros, err := repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener libros: %w", err)
	}
	return libros, nil
}

func DeleteLibro(ctx context.Context, repo LibroRepository, id int32) error {
	if _, err := GetLibro(ctx, repo, id); err != nil {
		return err
	}

	err := repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error al eliminar libro: %w", err)
	}
	return nil
}

func validarCronologiaLibro(ctx context.Context, autores AutorRepository, idAutor int32, lanzamiento Date) error {
	autor, err := autores.FindById(ctx, idAutor)
	if err != nil {
		if errors.Is(err, ErrAutorNoEncontrado) {
			return ErrReferencia
		}
		return err
	}
	minima := autor.FechaNacimiento.AddDate(5, 0, 0)
	// Para nacimientos del 29 de febrero, el aniversario en años no bisiestos
	// se considera el último día de febrero, igual que el intervalo de PostgreSQL.
	if autor.FechaNacimiento.Month() == time.February && autor.FechaNacimiento.Day() == 29 && minima.Month() == time.March {
		minima = minima.AddDate(0, 0, -1)
	}
	if lanzamiento.Format("2006-01-02") < minima.Format("2006-01-02") {
		return &ErrorValidacion{Mensaje: "El autor debe tener al menos 5 años cumplidos en la fecha de lanzamiento"}
	}
	return nil
}
