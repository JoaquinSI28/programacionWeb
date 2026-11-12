package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
	"bookfly.com/repository"
	"bookfly.com/views"
	"github.com/a-h/templ"
)

type webApp struct {
	libros      logic.LibroRepository
	autores     logic.AutorRepository
	editoriales logic.EditorialRepository
	usuarios    logic.UsuarioRepository
}

func newWebHandler(queries *dbsqlc.Queries) http.Handler {
	app := &webApp{
		libros: repository.NewLibroSQL(queries), autores: repository.NewAutorSQL(queries),
		editoriales: repository.NewEditorialSQL(queries), usuarios: repository.NewUsuarioSQL(queries),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.index)
	mux.HandleFunc("GET /estado", app.estado)
	mux.HandleFunc("GET /salud", func(w http.ResponseWriter, r *http.Request) {
		if _, err := queries.ComprobarConexion(r.Context()); err != nil {
			http.Error(w, "Base no disponible", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	for _, entity := range []string{"libros", "autores", "editoriales", "usuarios"} {
		mux.HandleFunc("POST /"+entity, func(w http.ResponseWriter, r *http.Request) {
			app.create(w, r, entity)
		})
		mux.HandleFunc("DELETE /"+entity+"/{id}", func(w http.ResponseWriter, r *http.Request) {
			app.delete(w, r, entity)
		})
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./API/static"))))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *webApp) pageData(ctx context.Context) (views.PageData, error) {
	data := views.PageData{Today: logic.Hoy()}
	var err error
	if data.Libros, err = logic.GetLibros(ctx, app.libros); err != nil {
		return data, err
	}
	if data.Autores, err = logic.GetAutores(ctx, app.autores); err != nil {
		return data, err
	}
	if data.Editoriales, err = logic.GetEditoriales(ctx, app.editoriales); err != nil {
		return data, err
	}
	if data.Usuarios, err = logic.GetUsuarios(ctx, app.usuarios); err != nil {
		return data, err
	}
	data.AutorNames = make(map[int32]string)
	for _, autor := range data.Autores {
		data.AutorNames[autor.IDAutor] = autor.Nombre
	}
	data.EditorNames = make(map[int32]string)
	for _, editorial := range data.Editoriales {
		data.EditorNames[editorial.IDEditorial] = editorial.Nombre
	}
	return data, nil
}

// Se renderiza antes de escribir cabeceras: un fallo no entrega una página a medias con 200.
func renderHTML(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	var page bytes.Buffer
	if err := component.Render(r.Context(), &page); err != nil {
		log.Printf("Error al renderizar la página: %v", err)
		http.Error(w, "No se pudo mostrar la página", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if _, err := page.WriteTo(w); err != nil {
		log.Printf("Error al enviar la página: %v", err)
	}
}

func (app *webApp) index(w http.ResponseWriter, r *http.Request) {
	data, err := app.pageData(r.Context())
	if err != nil {
		log.Printf("Error al consultar la biblioteca: %v", err)
		renderHTML(w, r, http.StatusInternalServerError, views.ErrorPage("No se pudo consultar la biblioteca. Intentá nuevamente."))
		return
	}
	renderHTML(w, r, http.StatusOK, views.IndexPage(data))
}

func parseForm(w http.ResponseWriter, r *http.Request) error {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/x-www-form-urlencoded" {
		return fmt.Errorf("el formulario debe enviarse como application/x-www-form-urlencoded")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("no se pudo leer el formulario")
	}
	return nil
}

func (app *webApp) formError(w http.ResponseWriter, r *http.Request, entity string, values url.Values, status int, message string) {
	target := "#mensajes"
	if entity != "" {
		target = "#" + entity + "-mensajes"
	}
	if r.Method == http.MethodDelete {
		if id, err := positiveNumber(r.PathValue("id"), "El ID"); err == nil {
			target = fmt.Sprintf("#%s-error-%d", entity, id)
		}
	}
	w.Header().Set("HX-Retarget", target)
	w.Header().Set("HX-Reswap", "innerHTML")
	renderHTML(w, r, status, views.OperationMessage(message))
}

func classifyError(err error, entity string) (int, string) {
	var validation *logic.ErrorValidacion
	switch {
	case errors.As(err, &validation):
		return http.StatusBadRequest, validation.Mensaje
	case errors.Is(err, logic.ErrDuplicado):
		switch entity {
		case "autores":
			return http.StatusConflict, "Ya existe un autor con ese nombre, país y fecha de nacimiento."
		case "editoriales":
			return http.StatusConflict, "Ya existe una editorial con ese nombre y sede."
		case "libros":
			return http.StatusConflict, "Ya existe esa edición del libro para el autor, la editorial y la fecha indicados."
		default:
			return http.StatusConflict, "El nombre de usuario o el correo ya está registrado."
		}
	case errors.Is(err, logic.ErrReferencia):
		return http.StatusConflict, "El registro referenciado no existe o tiene datos relacionados."
	case errors.Is(err, logic.ErrLibroNoEncontrado), errors.Is(err, logic.ErrAutorNoEncontrado),
		errors.Is(err, logic.ErrEditorialNoEncontrada), errors.Is(err, logic.ErrUsuarioNoEncontrado):
		return http.StatusNotFound, "El registro solicitado no existe."
	default:
		log.Printf("Error al modificar la biblioteca: %v", err)
		return http.StatusInternalServerError, "No se pudo completar la operación. Intentá nuevamente."
	}
}

func positiveNumber(value, label string) (int32, error) {
	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil || number <= 0 {
		return 0, &logic.ErrorValidacion{Mensaje: label + " debe ser un número entero positivo"}
	}
	return int32(number), nil
}

func formDate(value, label string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, &logic.ErrorValidacion{Mensaje: label + " debe tener formato YYYY-MM-DD"}
	}
	return date, nil
}

func (app *webApp) create(w http.ResponseWriter, r *http.Request, entity string) {
	if err := parseForm(w, r); err != nil {
		app.formError(w, r, entity, nil, http.StatusBadRequest, err.Error())
		return
	}
	values := r.PostForm
	for key := range values {
		if key != "contrasena" {
			values.Set(key, strings.TrimSpace(values.Get(key)))
		}
	}
	var err error
	switch entity {
	case "editoriales":
		_, err = logic.CreateEditorial(r.Context(), app.editoriales, values.Get("nombre"), values.Get("sede"))
	case "usuarios":
		_, err = logic.CreateUsuario(r.Context(), app.usuarios, values.Get("nombre"), values.Get("email"), values.Get("contrasena"))
	case "autores":
		var date time.Time
		date, err = formDate(values.Get("fecha_nacimiento"), "La fecha de nacimiento")
		if err == nil {
			_, err = logic.CreateAutor(r.Context(), app.autores, values.Get("nombre"), values.Get("pais"), date, values.Get("biografia"))
		}
	case "libros":
		err = app.createLibro(r.Context(), values)
	}
	if err != nil {
		status, message := classifyError(err, entity)
		app.formError(w, r, entity, values, status, message)
		return
	}
	data, err := app.pageData(r.Context())
	if err != nil {
		app.formError(w, r, entity, nil, http.StatusInternalServerError, "El registro se guardó, pero no se pudo actualizar la lista. Recargá la página para comprobarlo.")
		return
	}
	w.Header().Set("HX-Trigger", "catalogoActualizado")
	renderHTML(w, r, http.StatusOK, views.EntityList(entity, data))
}

func (app *webApp) createLibro(ctx context.Context, values url.Values) error {
	autor, err := positiveNumber(values.Get("id_autor"), "El autor")
	if err != nil {
		return err
	}
	editorial, err := positiveNumber(values.Get("id_editorial"), "La editorial")
	if err != nil {
		return err
	}
	pages, err := positiveNumber(values.Get("cantidad_paginas"), "La cantidad de páginas")
	if err != nil {
		return err
	}
	date, err := formDate(values.Get("fecha_lanzamiento"), "La fecha de lanzamiento")
	if err != nil {
		return err
	}
	_, err = logic.CreateLibro(ctx, app.libros, app.autores, autor, values.Get("titulo"), editorial, date, pages)
	return err
}

// La respuesta vacía permite que HTMX retire únicamente el elemento borrado.
func (app *webApp) delete(w http.ResponseWriter, r *http.Request, entity string) {
	id, err := positiveNumber(r.PathValue("id"), "El ID")
	if err == nil {
		switch entity {
		case "libros":
			err = logic.DeleteLibro(r.Context(), app.libros, id)
		case "autores":
			err = logic.DeleteAutor(r.Context(), app.autores, id)
		case "editoriales":
			err = logic.DeleteEditorial(r.Context(), app.editoriales, id)
		case "usuarios":
			err = logic.DeleteUsuario(r.Context(), app.usuarios, id)
		}
	}
	if err != nil {
		status, message := classifyError(err, entity)
		if errors.Is(err, logic.ErrReferencia) {
			switch entity {
			case "autores":
				message = "No se puede eliminar este autor porque tiene libros asociados. Eliminá primero esos libros."
			case "editoriales":
				message = "No se puede eliminar esta editorial porque tiene libros asociados. Eliminá primero esos libros."
			default:
				message = "No se puede eliminar este registro porque tiene lecturas asociadas."
			}
		}
		app.formError(w, r, entity, nil, status, message)
		return
	}
	w.Header().Set("HX-Trigger", "catalogoActualizado")
	w.WriteHeader(http.StatusOK)
}

// Refresca contadores y referencias sin reemplazar los borradores de formularios.
func (app *webApp) estado(w http.ResponseWriter, r *http.Request) {
	data, err := app.pageData(r.Context())
	if err != nil {
		app.formError(w, r, "", nil, http.StatusInternalServerError, "No se pudo actualizar el estado de la biblioteca.")
		return
	}
	data.ActiveForm = "libros"
	data.Values = map[string]string{"id_autor": r.URL.Query().Get("id_autor"), "id_editorial": r.URL.Query().Get("id_editorial")}
	renderHTML(w, r, http.StatusOK, views.CatalogState(data))
}
