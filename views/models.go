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
	ActiveForm  string
	Values      map[string]string
	Error       string
	Today       string
}

func (p PageData) Value(entity, field string) string {
	if entity != p.ActiveForm || field == "contrasena" {
		return ""
	}
	return p.Values[field]
}

func idString(id int32) string { return strconv.FormatInt(int64(id), 10) }

func countText(count int) string {
	if count == 1 {
		return "1 registro"
	}
	return strconv.Itoa(count) + " registros"
}
