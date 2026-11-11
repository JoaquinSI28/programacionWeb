package repository

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
)

func hashContrasena(password string) (string, error) {
	const iteraciones = 600000
	salt := make([]byte, 16)
	rand.Read(salt)
	hash, err := pbkdf2.Key(sha256.New, password, salt, iteraciones, 32)
	if err != nil {
		return "", err
	}
	encode := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iteraciones, encode(salt), encode(hash)), nil
}

// Los volúmenes creados con el TP4 guardan la contraseña en la columna anterior.
// La transacción permite reintentar una migración interrumpida sin perder usuarios.
func MigrarContrasenas(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "LOCK TABLE usuario IN ACCESS EXCLUSIVE MODE"); err != nil {
		return err
	}
	var anterior bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM pg_attribute WHERE attrelid = 'usuario'::regclass AND attname = 'contrasena' AND NOT attisdropped)").Scan(&anterior); err != nil {
		return err
	}
	if !anterior {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, "ALTER TABLE usuario ADD COLUMN contrasena_hash varchar(255)"); err != nil {
		return err
	}
	for {
		var id int32
		var password string
		err := tx.QueryRowContext(ctx, "SELECT id_usuario, contrasena FROM usuario WHERE contrasena_hash IS NULL LIMIT 1").Scan(&id, &password)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		hash, err := hashContrasena(password)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE usuario SET contrasena_hash = $1 WHERE id_usuario = $2", hash, id); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "ALTER TABLE usuario ALTER COLUMN contrasena_hash SET NOT NULL, DROP COLUMN contrasena"); err != nil {
		return err
	}
	return tx.Commit()
}
