package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
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
	if utf8.RuneCountInString(titulo) > 255 {
		return &ErrorValidacion{Mensaje: "El texto supera la longitud permitida para el campo."}
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

func validarCronologiaLibro(ctx context.Context, autores AutorRepository, idAutor int32, lanzamiento Date) error {
	autor, err := autores.FindById(ctx, idAutor)
	if errors.Is(err, ErrAutorNoEncontrado) {
		return ErrReferencia
	}
	if err != nil {
		return err
	}
	return ValidarEdadAlLanzamiento(autor.FechaNacimiento, lanzamiento)
}

func CreateLibro(ctx context.Context, uow UnitOfWork, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error) {
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

	var creado *Libro
	err := uow.Execute(ctx, func(store RepositoryProvider) error {
		if err := validarCronologiaLibro(ctx, store.Autores(), idAutor, fechaLanzamiento); err != nil {
			return err
		}
		var err error
		creado, err = store.Libros().Create(ctx, idAutor, titulo, idEditorial, fechaLanzamiento, cantidadPaginas)
		return err
	})
	if err != nil {
		return nil, err
	}
	return creado, nil
}

func UpdateLibro(ctx context.Context, uow UnitOfWork, id int32, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error) {
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

	var actualizado *Libro
	err := uow.Execute(ctx, func(store RepositoryProvider) error {
		if err := validarCronologiaLibro(ctx, store.Autores(), idAutor, fechaLanzamiento); err != nil {
			return err
		}
		var err error
		actualizado, err = store.Libros().Update(ctx, id, idAutor, titulo, idEditorial, fechaLanzamiento, cantidadPaginas)
		return err
	})
	if err != nil {
		return nil, err
	}
	return actualizado, nil
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
	err := repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error al eliminar libro: %w", err)
	}
	return nil
}
