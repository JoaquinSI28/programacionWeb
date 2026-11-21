package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Autor struct {
	IDAutor         int32
	Nombre          string
	Pais            string
	FechaNacimiento Date
	Biografia       string
}

var ErrAutorNoEncontrado = errors.New("autor no encontrado")

func validateAutor(nombre, pais, biografia string, fecha Date) error {
	if strings.TrimSpace(nombre) == "" {
		return &ErrorValidacion{Mensaje: "El nombre no puede estar vacío"}
	}
	if utf8.RuneCountInString(nombre) > 255 {
		return &ErrorValidacion{Mensaje: "El texto supera la longitud permitida para el campo."}
	}
	if strings.TrimSpace(pais) == "" {
		return &ErrorValidacion{Mensaje: "El país no puede estar vacío"}
	}
	if utf8.RuneCountInString(pais) > 100 {
		return &ErrorValidacion{Mensaje: "El texto supera la longitud permitida para el campo."}
	}
	if strings.TrimSpace(biografia) == "" {
		return &ErrorValidacion{Mensaje: "La biografía no puede estar vacía"}
	}
	return validarFecha(fecha, "La fecha de nacimiento")
}

func validarAutorDuplicado(ctx context.Context, repo AutorRepository, idActual int32, nombre, pais string, fechaNacimiento Date) error {
	autores, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener autores: %w", err)
	}
	claveNombre := normalizarClave(nombre)
	clavePais := normalizarClave(pais)
	fecha := fechaNacimiento.Format(time.DateOnly)
	for _, autor := range autores {
		if autor.IDAutor != idActual &&
			normalizarClave(autor.Nombre) == claveNombre &&
			normalizarClave(autor.Pais) == clavePais &&
			autor.FechaNacimiento.Format(time.DateOnly) == fecha {
			return ErrDuplicado
		}
	}
	return nil
}

func validarCronologiaLibrosDeAutor(ctx context.Context, libros LibroRepository, idAutor int32, fechaNacimiento Date) error {
	if libros == nil {
		return nil
	}
	lista, err := libros.List(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener libros: %w", err)
	}
	minima := fechaMinimaPublicacion(fechaNacimiento).Format(time.DateOnly)
	for _, libro := range lista {
		if libro.IDAutor == idAutor && libro.FechaLanzamiento.Format(time.DateOnly) < minima {
			return &ErrorValidacion{Mensaje: "El autor debe tener al menos 5 años cumplidos al lanzamiento de todos sus libros"}
		}
	}
	return nil
}

func CreateAutor(ctx context.Context, repo AutorRepository, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error) {
	if err := validateAutor(nombre, pais, biografia, fechaNacimiento); err != nil {
		return nil, fmt.Errorf("error al validar autor: %w", err)
	}
	if err := validarAutorDuplicado(ctx, repo, 0, nombre, pais, fechaNacimiento); err != nil {
		return nil, err
	}
	return repo.Create(ctx, nombre, pais, fechaNacimiento, biografia)
}

func GetAutor(ctx context.Context, repo AutorRepository, id int32) (*Autor, error) {
	autor, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener autor: %w", err)
	}
	return autor, nil
}

func GetAutores(ctx context.Context, repo AutorRepository) ([]*Autor, error) {
	autores, err := repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener autores: %w", err)
	}
	return autores, nil
}

func UpdateAutor(ctx context.Context, repo AutorRepository, libros LibroRepository, id int32, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error) {
	if err := validateAutor(nombre, pais, biografia, fechaNacimiento); err != nil {
		return nil, fmt.Errorf("error al validar autor: %w", err)
	}
	if _, err := GetAutor(ctx, repo, id); err != nil {
		return nil, err
	}
	if err := validarAutorDuplicado(ctx, repo, id, nombre, pais, fechaNacimiento); err != nil {
		return nil, err
	}
	if err := validarCronologiaLibrosDeAutor(ctx, libros, id, fechaNacimiento); err != nil {
		return nil, err
	}
	return repo.Update(ctx, id, nombre, pais, fechaNacimiento, biografia)
}

func DeleteAutor(ctx context.Context, repo AutorRepository, libros LibroRepository, id int32) error {
	if _, err := GetAutor(ctx, repo, id); err != nil {
		return err
	}
	if libros != nil {
		lista, err := libros.List(ctx)
		if err != nil {
			return fmt.Errorf("error al obtener libros: %w", err)
		}
		for _, libro := range lista {
			if libro.IDAutor == id {
				return ErrReferencia
			}
		}
	}
	if err := repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error al eliminar autor: %w", err)
	}
	return nil
}
