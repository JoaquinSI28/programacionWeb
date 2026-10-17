package logic

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Usuario struct {
	ID         int32
	Nombre     string
	Email      string
	Contrasena string
}

var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")

func validateNombre(nombre string) error {
	if strings.TrimSpace(nombre) == "" {
		return &ErrorValidacion{Mensaje: "El nombre no puede estar vacío"}
	}
	return nil
}
func validateEmail(email string) error {
	if email == "" {
		return &ErrorValidacion{Mensaje: "El email no puede estar vacío"}
	}

	re := regexp.MustCompile(
		`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !re.MatchString(email) {
		return &ErrorValidacion{Mensaje: "formato de email inválido"}
	}

	return nil
}
func validateContrasena(contrasena string) error {
	if contrasena == "" {
		return &ErrorValidacion{Mensaje: "La contraseña no puede estar vacía"}
	}
	if len(contrasena) < 3 {
		return &ErrorValidacion{Mensaje: "La contraseña debe tener al menos 3 caracteres"}
	}
	return nil
}

func CreateUsuario(ctx context.Context, repo UsuarioRepository, nombre string, email string, contrasena string) (*Usuario, error) {
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("error al validar email: %w", err)
	}
	if err := validateContrasena(contrasena); err != nil {
		return nil, fmt.Errorf("error al validar contraseña: %w", err)
	}
	if err := validateNombre(nombre); err != nil {
		return nil, fmt.Errorf("error al validar nombre: %w", err)
	}

	return repo.Create(ctx, nombre, email, contrasena)
}

func GetUsuario(ctx context.Context, repo UsuarioRepository, id int32) (*Usuario, error) {
	usuario, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}
	return usuario, nil
}
func GetUsuarios(ctx context.Context, repo UsuarioRepository) ([]*Usuario, error) {
	usuarios, err := repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuarios: %w", err)
	}
	return usuarios, nil
}

func UpdateUsuario(ctx context.Context, repo UsuarioRepository, id int32, nombre, email, contrasena string) (*Usuario, error) {
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("error al validar email: %w", err)
	}
	if err := validateContrasena(contrasena); err != nil {
		return nil, fmt.Errorf("error al validar contraseña: %w", err)
	}
	if err := validateNombre(nombre); err != nil {
		return nil, fmt.Errorf("error al validar nombre: %w", err)
	}

	_, err := repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al buscar usuario para actualizar: %w", err)
	}

	return repo.Update(ctx, id, nombre, email, contrasena)
}

func DeleteUsuario(ctx context.Context, repo UsuarioRepository, id int32) error {
	if _, err := GetUsuario(ctx, repo, id); err != nil {
		return err
	}

	err := repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("error al eliminar usuario: %w", err)
	}
	return nil
}
