package main

import (
	"bytes"
	"context"
	"database/sql"
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

func newWebHandler(conn *sql.DB) http.Handler {
	queries := dbsqlc.New(conn)
	app := &webApp{
		libros: repository.NewLibroSQL(conn), autores: repository.NewAutorSQL(conn),
		editoriales: repository.NewEditorialSQL(queries), usuarios: repository.NewUsuarioSQL(queries),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.index)
	mux.HandleFunc("GET /salud", func(w http.ResponseWriter, r *http.Request) {
		if _, err := queries.ComprobarConexion(r.Context()); err != nil {
			http.Error(w, "Base no disponible", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /opciones/autores", app.autorOptions)
	mux.HandleFunc("GET /opciones/editoriales", app.editorialOptions)
	for _, entity := range []string{"libros", "autores", "editoriales", "usuarios"} {
		mux.HandleFunc("GET /formularios/"+entity, func(w http.ResponseWriter, r *http.Request) {
			app.form(w, r, entity)
		})
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

// Cada fragmento consulta únicamente las colecciones que va a mostrar.
func (app *webApp) pageData(ctx context.Context, entities ...string) (views.PageData, error) {
	data := views.PageData{Today: logic.Hoy()}
	for _, entity := range entities {
		var err error
		switch entity {
		case "libros":
			data.Libros, err = logic.GetLibros(ctx, app.libros)
		case "autores":
			data.Autores, err = logic.GetAutores(ctx, app.autores)
		case "editoriales":
			data.Editoriales, err = logic.GetEditoriales(ctx, app.editoriales)
		case "usuarios":
			data.Usuarios, err = logic.GetUsuarios(ctx, app.usuarios)
		}
		if err != nil {
			return data, err
		}
	}
	if len(data.Libros) > 0 {
		data.AutorNames = make(map[int32]string, len(data.Autores))
		for _, autor := range data.Autores {
			data.AutorNames[autor.IDAutor] = autor.Nombre
		}
		data.EditorNames = make(map[int32]string, len(data.Editoriales))
		for _, editorial := range data.Editoriales {
			data.EditorNames[editorial.IDEditorial] = editorial.Nombre
		}
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
	data, err := app.pageData(r.Context(), "libros", "autores", "editoriales", "usuarios")
	if err != nil {
		log.Printf("Error al consultar la biblioteca: %v", err)
		renderHTML(w, r, http.StatusInternalServerError, views.ErrorPage("No se pudo consultar la biblioteca. Intentá nuevamente."))
		return
	}
	renderHTML(w, r, http.StatusOK, views.IndexPage(data))
}

// El alta refresca sólo su propio formulario. En libros también se cargan sus referencias.
func (app *webApp) form(w http.ResponseWriter, r *http.Request, entity string) {
	data := views.PageData{Today: logic.Hoy()}
	if entity == "libros" {
		var err error
		data, err = app.pageData(r.Context(), "autores", "editoriales")
		if err != nil {
			operationError(w, r, "libros-form-error", http.StatusInternalServerError, "No se pudo actualizar el formulario de libros. Recargá la página.")
			return
		}
	}
	renderHTML(w, r, http.StatusOK, views.EntityForm(entity, data))
}

func (app *webApp) autorOptions(w http.ResponseWriter, r *http.Request) {
	autores, err := logic.GetAutores(r.Context(), app.autores)
	if err != nil {
		operationError(w, r, "libros-form-error", http.StatusInternalServerError, "No se pudo actualizar el selector de autores. Recargá la página.")
		return
	}
	renderHTML(w, r, http.StatusOK, views.AutorOptions(autores, r.URL.Query().Get("id_autor")))
}

func (app *webApp) editorialOptions(w http.ResponseWriter, r *http.Request) {
	editoriales, err := logic.GetEditoriales(r.Context(), app.editoriales)
	if err != nil {
		operationError(w, r, "libros-form-error", http.StatusInternalServerError, "No se pudo actualizar el selector de editoriales. Recargá la página.")
		return
	}
	renderHTML(w, r, http.StatusOK, views.EditorialOptions(editoriales, r.URL.Query().Get("id_editorial")))
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

func operationError(w http.ResponseWriter, r *http.Request, id string, status int, message string) {
	renderHTML(w, r, status, views.OperationMessage(id, message))
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
	errorTarget := entity + "-form-error"
	if err := parseForm(w, r); err != nil {
		operationError(w, r, errorTarget, http.StatusBadRequest, err.Error())
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
		operationError(w, r, errorTarget, status, message)
		return
	}
	collections := []string{entity}
	if entity == "libros" {
		collections = append(collections, "autores", "editoriales")
	}
	data, err := app.pageData(r.Context(), collections...)
	if err != nil {
		operationError(w, r, errorTarget, http.StatusInternalServerError, "El registro se guardó, pero no se pudo actualizar la lista. Recargá la página para comprobarlo.")
		return
	}
	renderHTML(w, r, http.StatusOK, views.EntityList(entity, data, true))
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
	_, err = logic.CreateLibro(ctx, app.libros, autor, values.Get("titulo"), editorial, date, pages)
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
		operationError(w, r, entity+"-"+strconv.FormatInt(int64(id), 10)+"-error", status, message)
		return
	}
	w.WriteHeader(http.StatusOK)
}
