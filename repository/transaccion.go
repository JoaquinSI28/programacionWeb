package repository

import (
	"context"
	"database/sql"
	"errors"

	dbsqlc "bookfly.com/db/sqlc"
	"github.com/lib/pq"
)

// Una publicación y un cambio de nacimiento pueden leer los mismos datos.
// Si PostgreSQL detecta ese conflicto, se vuelve a validar con los datos actuales.
func conTransaccionSerializable(ctx context.Context, conn *sql.DB, escribir func(*dbsqlc.Queries) error) error {
	const maxIntentos = 3
	for intento := 1; ; intento++ {
		tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}
		if err = escribir(dbsqlc.New(tx)); err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
		var pgErr *pq.Error
		if intento == maxIntentos || !errors.As(err, &pgErr) || pgErr.Code != "40001" {
			return err
		}
	}
}
