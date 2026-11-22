package views

import (
	"strconv"

	"bookfly.com/logic"
)

// PageData contiene los datos que necesita la página completa, sin depender de sqlc.
type PageData struct {
	Libros      []*logic.Libro
	Autores     []*logic.Autor
	Editoriales []*logic.Editorial
	Usuarios    []*logic.Usuario
	AutorNames  map[int32]string
	EditorNames map[int32]string
	Today       string
}

func idString(id int32) string { return strconv.FormatInt(int64(id), 10) }
