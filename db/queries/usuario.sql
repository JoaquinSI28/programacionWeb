-- name: CreateUsuario :one
INSERT INTO Usuario (nombre_usuario, mail, contrasena) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetUsuario :one
SELECT FROM Usuario 
WHERE id_usuario = $1;

-- name: ListUsuarios :many
SELECT * FROM Usuario 
ORDER BY nombre_usuario;

-- name: UpdateUsuario :exec
UPDATE Usuario 
SET nombre_usuario = $2, mail = $3, contrasena = $4
WHERE id_usuario = $1;

-- name: DeleteUsuario :exec
DELETE FROM Usuario 
WHERE id_usuario = $1;
