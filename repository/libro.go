package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
)

// LibroSQL implementa logic.LibroRepository usando las consultas de sqlc.
type LibroSQL struct {
	queries *dbsqlc.Queries
}

func NewLibroSQL(queries *dbsqlc.Queries) *LibroSQL {
	return &LibroSQL{queries: queries}
}

func (r *LibroSQL) Create(ctx context.Context, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento logic.Date, cantidadPaginas int32) (*logic.Libro, error) {
	creado, err := r.queries.CreateLibro(ctx, dbsqlc.CreateLibroParams{
		Titulo:           titulo,
		IDAutor:          idAutor,
		IDEditorial:      idEditorial,
		FechaLanzamiento: fechaLanzamiento,
		CantidadPaginas:  cantidadPaginas,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Libro{
		IDLibro:          creado.IDLibro,
		Titulo:           creado.Titulo,
		IDAutor:          creado.IDAutor,
		IDEditorial:      creado.IDEditorial,
		FechaLanzamiento: creado.FechaLanzamiento,
		CantidadPaginas:  creado.CantidadPaginas,
	}, nil
}

func (r *LibroSQL) FindById(ctx context.Context, id int32) (*logic.Libro, error) {
	libro, err := r.queries.GetLibro(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, logic.ErrLibroNoEncontrado
		}
		return nil, err
	}

	return &logic.Libro{
		IDLibro:          libro.IDLibro,
		Titulo:           libro.Titulo,
		IDAutor:          libro.IDAutor,
		IDEditorial:      libro.IDEditorial,
		FechaLanzamiento: libro.FechaLanzamiento,
		CantidadPaginas:  libro.CantidadPaginas,
	}, nil
}

func (r *LibroSQL) List(ctx context.Context) ([]*logic.Libro, error) {
	libros, err := r.queries.ListLibros(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*logic.Libro, len(libros))
	for i, libro := range libros {
		result[i] = &logic.Libro{
			IDLibro:          libro.IDLibro,
			Titulo:           libro.Titulo,
			IDAutor:          libro.IDAutor,
			IDEditorial:      libro.IDEditorial,
			FechaLanzamiento: libro.FechaLanzamiento,
			CantidadPaginas:  libro.CantidadPaginas,
		}
	}

	return result, nil
}

func (r *LibroSQL) Delete(ctx context.Context, id int32) error {
	err := r.queries.DeleteLibro(ctx, id)
	if err != nil {
		return errorPersistencia(err)
	}
	return nil
}

func (r *LibroSQL) Update(ctx context.Context, id int32, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento logic.Date, cantidadPaginas int32) (*logic.Libro, error) {
	err := r.queries.UpdateLibro(ctx, dbsqlc.UpdateLibroParams{
		IDLibro:          id,
		Titulo:           titulo,
		IDAutor:          idAutor,
		IDEditorial:      idEditorial,
		FechaLanzamiento: fechaLanzamiento,
		CantidadPaginas:  cantidadPaginas,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Libro{
		IDLibro:          id,
		Titulo:           titulo,
		IDAutor:          idAutor,
		IDEditorial:      idEditorial,
		FechaLanzamiento: fechaLanzamiento,
		CantidadPaginas:  cantidadPaginas,
	}, nil
}

var _ logic.LibroRepository = (*LibroSQL)(nil)
