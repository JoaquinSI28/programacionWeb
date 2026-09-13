-- name: CreateLee :one
INSERT INTO Lee (id_usuario, id_libro, estado, resena) 
VALUES ($1, $2, $3, $4) 
RETURNING *;

-- name: GetLee :one
SELECT FROM Lee 
WHERE id_usuario = $1 AND id_libro = $2;

-- name: ListLees :many
SELECT * FROM Lee;

-- name: UpdateLee :exec
UPDATE Lee 
SET estado = $3, resena = $4
WHERE id_usuario = $1 AND id_libro = $2;

-- name: DeleteLee :exec
DELETE FROM Lee 
WHERE id_usuario = $1 AND id_libro = $2;
