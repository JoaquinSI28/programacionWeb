-- name: CreateUsuario :one
INSERT INTO Usuario (nombre_usuario, mail, contrasena_hash)
VALUES ($1, $2, $3)
RETURNING id_usuario, nombre_usuario, mail;

-- name: GetUsuario :one
SELECT id_usuario, nombre_usuario, mail FROM Usuario
WHERE id_usuario = $1;

-- name: ListUsuarios :many
SELECT id_usuario, nombre_usuario, mail FROM Usuario
ORDER BY nombre_usuario;

-- name: UpdateUsuario :exec
UPDATE Usuario
SET nombre_usuario = $2, mail = $3, contrasena_hash = $4
WHERE id_usuario = $1;

-- name: DeleteUsuario :exec
DELETE FROM Usuario
WHERE id_usuario = $1;
