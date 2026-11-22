-- name: CreateEditorial :one
INSERT INTO Editorial (nombre, sede) 
VALUES ($1, $2) 
RETURNING *;

-- name: GetEditorial :one
SELECT * FROM Editorial 
WHERE id_editorial = $1;

-- name: ListEditoriales :many
SELECT * FROM Editorial 
ORDER BY nombre;

-- name: UpdateEditorial :one
UPDATE Editorial 
SET nombre = $2, sede = $3
WHERE id_editorial = $1
RETURNING id_editorial;

-- name: DeleteEditorial :one
DELETE FROM Editorial 
WHERE id_editorial = $1
RETURNING id_editorial;
