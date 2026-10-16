package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Autor struct {
	IDAutor         int32
	Nombre          string
	Pais            string
	FechaNacimiento Date
	Biografia       string
}

var ErrAutorNoEncontrado = errors.New("autor no encontrado")

func validateAutor(nombre, pais, biografia string) error {
	if strings.TrimSpace(nombre) == "" {
		return &ErrorValidacion{Mensaje: "El nombre no puede estar vacío"}
	}
	if strings.TrimSpace(pais) == "" {
		return &ErrorValidacion{Mensaje: "El país no puede estar vacío"}
	}
	if strings.TrimSpace(biografia) == "" {
		return &ErrorValidacion{Mensaje: "La biografía no puede estar vacía"}
	}
	return nil
}

func CreateAutor(ctx context.Context, repo AutorRepository, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error) {
	if err := validateAutor(nombre, pais, biografia); err != nil {
		return nil, fmt.Errorf("error al validar autor: %w", err)
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

func UpdateAutor(ctx context.Context, repo AutorRepository, id int32, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error) {
	if err := validateAutor(nombre, pais, biografia); err != nil {
		return nil, fmt.Errorf("error al validar autor: %w", err)
	}
	if _, err := GetAutor(ctx, repo, id); err != nil {
		return nil, err
	}
	return repo.Update(ctx, id, nombre, pais, fechaNacimiento, biografia)
}

func DeleteAutor(ctx context.Context, repo AutorRepository, id int32) error {
	if _, err := GetAutor(ctx, repo, id); err != nil {
		return err
	}
	if err := repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error al eliminar autor: %w", err)
	}
	return nil
}
