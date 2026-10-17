package repository

import (
	"errors"

	"bookfly.com/logic"
	"github.com/lib/pq"
)

func errorPersistencia(err error) error {
	var postgres *pq.Error
	if errors.As(err, &postgres) {
		switch postgres.Code {
		case "23505":
			return logic.ErrDuplicado
		case "23503":
			return logic.ErrReferencia
		}
	}
	return err
}
