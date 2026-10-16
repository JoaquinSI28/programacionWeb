package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Editorial struct {
	IDEditorial int32
	Nombre      string
	Sede        string
}

var ErrEditorialNoEncontrada = errors.New("editorial no encontrada")

func validateEditorial(nombre, sede string) error {
	if strings.TrimSpace(nombre) == "" {
		return &ErrorValidacion{Mensaje: "El nombre no puede estar vacío"}
	}
	if strings.TrimSpace(sede) == "" {
		return &ErrorValidacion{Mensaje: "La sede no puede estar vacía"}
	}
	return nil
}

func CreateEditorial(ctx context.Context, repo EditorialRepository, nombre, sede string) (*Editorial, error) {
	if err := validateEditorial(nombre, sede); err != nil {
		return nil, fmt.Errorf("error al validar editorial: %w", err)
	}
	return repo.Create(ctx, nombre, sede)
}

func GetEditorial(ctx context.Context, repo EditorialRepository, id int32) (*Editorial, error) {
	editorial, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener editorial: %w", err)
	}
	return editorial, nil
}

func GetEditoriales(ctx context.Context, repo EditorialRepository) ([]*Editorial, error) {
	editoriales, err := repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener editoriales: %w", err)
	}
	return editoriales, nil
}

func UpdateEditorial(ctx context.Context, repo EditorialRepository, id int32, nombre, sede string) (*Editorial, error) {
	if err := validateEditorial(nombre, sede); err != nil {
		return nil, fmt.Errorf("error al validar editorial: %w", err)
	}
	if _, err := GetEditorial(ctx, repo, id); err != nil {
		return nil, err
	}
	return repo.Update(ctx, id, nombre, sede)
}

func DeleteEditorial(ctx context.Context, repo EditorialRepository, id int32) error {
	if _, err := GetEditorial(ctx, repo, id); err != nil {
		return err
	}
	if err := repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error al eliminar editorial: %w", err)
	}
	return nil
}
