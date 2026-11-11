package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
)

// UsuarioSQL implementa logic.UsuarioRepository usando las consultas de sqlc.
type UsuarioSQL struct {
	queries *dbsqlc.Queries
}

func (r *UsuarioSQL) FindById(ctx context.Context, id int32) (*logic.Usuario, error) {
	usuario, err := r.queries.GetUsuario(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, logic.ErrUsuarioNoEncontrado
		}
		return nil, err
	}

	return &logic.Usuario{
		ID:     usuario.IDUsuario,
		Nombre: usuario.NombreUsuario,
		Email:  usuario.Mail,
	}, nil
}

func NewUsuarioSQL(queries *dbsqlc.Queries) *UsuarioSQL {
	return &UsuarioSQL{queries: queries}
}

func (r *UsuarioSQL) Create(ctx context.Context, nombre, email, contrasena string) (*logic.Usuario, error) {
	hash, err := hashContrasena(contrasena)
	if err != nil {
		return nil, err
	}
	creado, err := r.queries.CreateUsuario(ctx, dbsqlc.CreateUsuarioParams{
		NombreUsuario:  nombre,
		Mail:           email,
		ContrasenaHash: hash,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Usuario{
		ID:     creado.IDUsuario,
		Nombre: creado.NombreUsuario,
		Email:  creado.Mail,
	}, nil
}
func (r *UsuarioSQL) List(ctx context.Context) ([]*logic.Usuario, error) {
	usuarios, err := r.queries.ListUsuarios(ctx)
	if err != nil {
		return nil, err
	}

	var result []*logic.Usuario
	for _, usuario := range usuarios {
		result = append(result, &logic.Usuario{
			ID:     usuario.IDUsuario,
			Nombre: usuario.NombreUsuario,
			Email:  usuario.Mail,
		})
	}

	return result, nil
}

func (r *UsuarioSQL) Delete(ctx context.Context, id int32) error {
	err := r.queries.DeleteUsuario(ctx, id)
	if err != nil {
		return errorPersistencia(err)
	}
	return nil
}

func (r *UsuarioSQL) Update(ctx context.Context, id int32, nombre, email, contrasena string) (*logic.Usuario, error) {
	hash, err := hashContrasena(contrasena)
	if err != nil {
		return nil, err
	}
	err = r.queries.UpdateUsuario(ctx, dbsqlc.UpdateUsuarioParams{
		IDUsuario:      id,
		NombreUsuario:  nombre,
		Mail:           email,
		ContrasenaHash: hash,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Usuario{
		ID:     id,
		Nombre: nombre,
		Email:  email,
	}, nil
}

var _ logic.UsuarioRepository = (*UsuarioSQL)(nil)
