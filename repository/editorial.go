package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
)

type EditorialSQL struct {
	queries *dbsqlc.Queries
}

func NewEditorialSQL(queries *dbsqlc.Queries) *EditorialSQL {
	return &EditorialSQL{queries: queries}
}

func (r *EditorialSQL) Create(ctx context.Context, nombre, sede string) (*logic.Editorial, error) {
	creado, err := r.queries.CreateEditorial(ctx, dbsqlc.CreateEditorialParams{
		Nombre: nombre,
		Sede:   sede,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Editorial{
		IDEditorial: creado.IDEditorial,
		Nombre:      creado.Nombre,
		Sede:        creado.Sede,
	}, nil
}

func (r *EditorialSQL) FindById(ctx context.Context, id int32) (*logic.Editorial, error) {
	editorial, err := r.queries.GetEditorial(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, logic.ErrEditorialNoEncontrada
		}
		return nil, err
	}

	return &logic.Editorial{
		IDEditorial: editorial.IDEditorial,
		Nombre:      editorial.Nombre,
		Sede:        editorial.Sede,
	}, nil
}

func (r *EditorialSQL) List(ctx context.Context) ([]*logic.Editorial, error) {
	editoriales, err := r.queries.ListEditoriales(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*logic.Editorial, len(editoriales))
	for i, editorial := range editoriales {
		result[i] = &logic.Editorial{
			IDEditorial: editorial.IDEditorial,
			Nombre:      editorial.Nombre,
			Sede:        editorial.Sede,
		}
	}
	return result, nil
}

func (r *EditorialSQL) Update(ctx context.Context, id int32, nombre, sede string) (*logic.Editorial, error) {
	err := r.queries.UpdateEditorial(ctx, dbsqlc.UpdateEditorialParams{
		IDEditorial: id,
		Nombre:      nombre,
		Sede:        sede,
	})
	if err != nil {
		return nil, errorPersistencia(err)
	}

	return &logic.Editorial{
		IDEditorial: id,
		Nombre:      nombre,
		Sede:        sede,
	}, nil
}

func (r *EditorialSQL) Delete(ctx context.Context, id int32) error {
	return errorPersistencia(r.queries.DeleteEditorial(ctx, id))
}

var _ logic.EditorialRepository = (*EditorialSQL)(nil)
