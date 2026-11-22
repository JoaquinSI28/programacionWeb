package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
)

type AutorSQL struct {
	conn    *sql.DB
	queries *dbsqlc.Queries
}

func NewAutorSQL(conn *sql.DB) *AutorSQL {
	return &AutorSQL{conn: conn, queries: dbsqlc.New(conn)}
}

func (r *AutorSQL) Create(ctx context.Context, nombre, pais string, fechaNacimiento logic.Date, biografia string) (*logic.Autor, error) {
	creado, err := r.queries.CreateAutor(ctx, dbsqlc.CreateAutorParams{
		Nombre:          nombre,
		Pais:            pais,
		FechaNacimiento: fechaNacimiento,
		Biografia:       biografia,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Autor{
		IDAutor:         creado.IDAutor,
		Nombre:          creado.Nombre,
		Pais:            creado.Pais,
		FechaNacimiento: creado.FechaNacimiento,
		Biografia:       creado.Biografia,
	}, nil
}

func (r *AutorSQL) FindById(ctx context.Context, id int32) (*logic.Autor, error) {
	autor, err := r.queries.GetAutor(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, logic.ErrAutorNoEncontrado
		}
		return nil, err
	}

	return &logic.Autor{
		IDAutor:         autor.IDAutor,
		Nombre:          autor.Nombre,
		Pais:            autor.Pais,
		FechaNacimiento: autor.FechaNacimiento,
		Biografia:       autor.Biografia,
	}, nil
}

func (r *AutorSQL) List(ctx context.Context) ([]*logic.Autor, error) {
	autores, err := r.queries.ListAutores(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*logic.Autor, len(autores))
	for i, autor := range autores {
		result[i] = &logic.Autor{
			IDAutor:         autor.IDAutor,
			Nombre:          autor.Nombre,
			Pais:            autor.Pais,
			FechaNacimiento: autor.FechaNacimiento,
			Biografia:       autor.Biografia,
		}
	}
	return result, nil
}

func (r *AutorSQL) Update(ctx context.Context, id int32, nombre, pais string, fechaNacimiento logic.Date, biografia string) (*logic.Autor, error) {
	err := conTransaccionSerializable(ctx, r.conn, func(q *dbsqlc.Queries) error {
		// Si el autor tenía edad suficiente para su primer libro, también para los posteriores.
		primero, err := q.PrimerLanzamientoDelAutor(ctx, id)
		switch {
		case err == nil:
			if err := logic.ValidarEdadAlLanzamiento(fechaNacimiento, primero); err != nil {
				return err
			}
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}
		_, err = q.UpdateAutor(ctx, dbsqlc.UpdateAutorParams{
			IDAutor: id, Nombre: nombre, Pais: pais, FechaNacimiento: fechaNacimiento, Biografia: biografia,
		})
		return errorDeEscritura(err, logic.ErrAutorNoEncontrado)
	})
	if err != nil {
		return nil, err
	}

	return &logic.Autor{
		IDAutor:         id,
		Nombre:          nombre,
		Pais:            pais,
		FechaNacimiento: fechaNacimiento,
		Biografia:       biografia,
	}, nil
}

func (r *AutorSQL) Delete(ctx context.Context, id int32) error {
	_, err := r.queries.DeleteAutor(ctx, id)
	return errorDeEscritura(err, logic.ErrAutorNoEncontrado)
}

var _ logic.AutorRepository = (*AutorSQL)(nil)
