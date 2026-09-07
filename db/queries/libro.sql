-- name: CreateLibro :one
INSERT INTO Libro (titulo, id_autor, id_editorial, fecha_lanzamiento, cantidad_paginas) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: GetLibro :one
SELECT FROM Libro 
WHERE id_libro = $1;

-- name: ListLibros :many
SELECT * FROM Libro 
ORDER BY titulo;

-- name: UpdateLibro :exec
UPDATE Libro 
SET titulo = $2, id_autor = $3, id_editorial = $4, fecha_lanzamiento = $5, cantidad_paginas = $6
WHERE id_libro = $1;

-- name: DeleteLibro :exec
DELETE FROM Libro 
WHERE id_libro = $1;
