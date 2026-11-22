package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

//go:embed reglas.sql
var reglas string

func AplicarReglas(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Dos instancias pueden iniciar al mismo tiempo sobre el mismo volumen.
	if _, err := tx.ExecContext(ctx, "LOCK TABLE autor, editorial, libro, usuario IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, reglas); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("hay registros duplicados en el catálogo anterior; revisá el índice indicado antes de reintentar: %w", err)
		}
		return err
	}
	return tx.Commit()
}
