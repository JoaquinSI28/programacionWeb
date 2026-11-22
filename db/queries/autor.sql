-- name: CreateAutor :one
INSERT INTO Autor (nombre, pais, fecha_nacimiento, biografia) 
VALUES ($1, $2, $3, $4) 
RETURNING *;

-- name: GetAutor :one
SELECT * FROM Autor 
WHERE id_autor = $1;

-- name: ListAutores :many
SELECT * FROM Autor 
ORDER BY nombre;

-- name: UpdateAutor :one
UPDATE Autor 
SET nombre = $2, pais = $3, fecha_nacimiento = $4, biografia = $5
WHERE id_autor = $1
RETURNING id_autor;

-- name: DeleteAutor :one
DELETE FROM Autor 
WHERE id_autor = $1
RETURNING id_autor;
