package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
	"github.com/lib/pq"
)

type sqlRepositoryProvider struct {
	queries *dbsqlc.Queries
}

func (p *sqlRepositoryProvider) Autores() logic.AutorRepository {
	return NewAutorSQL(p.queries)
}

func (p *sqlRepositoryProvider) Libros() logic.LibroRepository {
	return NewLibroSQL(p.queries)
}

// SQLUnitOfWork implementa logic.UnitOfWork mediante transacciones serializables de PostgreSQL.
// Una publicación y un cambio de nacimiento pueden leer los mismos datos; si PostgreSQL
// detecta ese conflicto (40001), se reintenta la unidad con los datos actualizados.
type SQLUnitOfWork struct {
	conn *sql.DB
}

func NewSQLUnitOfWork(conn *sql.DB) *SQLUnitOfWork {
	return &SQLUnitOfWork{conn: conn}
}

func (u *SQLUnitOfWork) Execute(ctx context.Context, fn func(logic.RepositoryProvider) error) error {
	const maxIntentos = 3
	for intento := 1; ; intento++ {
		tx, err := u.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}
		provider := &sqlRepositoryProvider{queries: dbsqlc.New(tx)}
		if err = fn(provider); err == nil {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
		var pgErr *pq.Error
		if intento == maxIntentos || !errors.As(err, &pgErr) || pgErr.Code != "40001" {
			return err
		}
	}
}

var _ logic.UnitOfWork = (*SQLUnitOfWork)(nil)
