package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	dbsqlc "bookfly.com/db/sqlc"
)

// Se ejecuta con make test sobre un PostgreSQL de prueba preparado por Compose.
func TestHTMLFormsAndPRG(t *testing.T) {
	t.Chdir("..")
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		envOrDefault("POSTGRES_USER", "postgres"), envOrDefault("POSTGRES_PASSWORD", "postgres"), envOrDefault("POSTGRES_DB", "prueba"))
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.Ping(); err != nil {
		t.Fatal(err)
	}
	queries := dbsqlc.New(conn)
	server := httptest.NewServer(newWebHandler(queries))
	defer server.Close()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ctx := context.Background()
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := "Editorial HTML " + suffix + " <script>alert(1)</script>"
	authorName, username := "Autor HTML "+suffix, "lector_"+suffix
	password := strings.Repeat("ñ", 255)
	title := "Libro HTML " + suffix
	var editorialID, authorID, bookID, userID int32
	defer func() {
		if bookID != 0 {
			_ = queries.DeleteLibro(ctx, bookID)
		}
		if authorID != 0 {
			_ = queries.DeleteAutor(ctx, authorID)
		}
		if editorialID != 0 {
			_ = queries.DeleteEditorial(ctx, editorialID)
		}
		if userID != 0 {
			_ = queries.DeleteUsuario(ctx, userID)
		}
	}()

	request := func(method, path, body, contentType string, want int) string {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: estado %d, esperado %d; %s", method, path, response.StatusCode, want, data)
		}
		if want == http.StatusSeeOther && response.Header.Get("Location") != "/" {
			t.Fatalf("redirección incorrecta: %q", response.Header.Get("Location"))
		}
		if want == http.StatusOK && path == "/" && !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") {
			t.Fatal("la página no es HTML")
		}
		return string(data)
	}
	post := func(path string, values url.Values, status int) string {
		t.Helper()
		return request(http.MethodPost, path, values.Encode(), "application/x-www-form-urlencoded", status)
	}
	get := func() string { t.Helper(); return request(http.MethodGet, "/", "", "", http.StatusOK) }
	if body := request(http.MethodGet, "/salud", "", "", http.StatusNoContent); body != "" {
		t.Fatal("salud devolvió contenido")
	}
	page := get()
	if !strings.Contains(page, "<!doctype html>") && !strings.Contains(page, "<!DOCTYPE html>") {
		t.Fatal("falta documento HTML completo")
	}
	if strings.Contains(page, "<script") || strings.Contains(page, "hx-post") {
		t.Fatal("la página requiere JavaScript o HTMX")
	}
	request(http.MethodGet, "/static/styles.css", "", "", http.StatusOK)
	request(http.MethodGet, "/app.js", "", "", http.StatusNotFound)
	post("/editoriales", url.Values{"nombre": {name}, "sede": {"Tandil"}}, http.StatusSeeOther)
	editoriales, err := queries.ListEditoriales(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range editoriales {
		if item.Nombre == name {
			editorialID = item.IDEditorial
		}
	}
	if editorialID == 0 {
		t.Fatal("el POST no persistió la editorial")
	}
	page = get()
	if strings.Contains(page, "<script>alert(1)</script>") || !strings.Contains(page, "&lt;script&gt;") {
		t.Fatal("el contenido del usuario no se escapó")
	}
	get() // Recargar el GET no debe repetir el POST.
	editoriales, _ = queries.ListEditoriales(ctx)
	count := 0
	for _, item := range editoriales {
		if item.Nombre == name {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("recargar creó %d editoriales", count)
	}

	post("/autores", url.Values{"nombre": {authorName}, "pais": {"Argentina"}, "fecha_nacimiento": {"1980-01-15"}, "biografia": {"Biografía"}}, http.StatusSeeOther)
	autores, err := queries.ListAutores(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range autores {
		if item.Nombre == authorName {
			authorID = item.IDAutor
		}
	}
	if authorID == 0 {
		t.Fatal("falta el autor creado")
	}
	userForm := url.Values{"nombre": {username}, "email": {username + "@example.com"}, "contrasena": {password}}
	userForm.Set("contrasena", password+"ñ")
	if body := post("/usuarios", userForm, http.StatusBadRequest); !strings.Contains(body, "255 caracteres") || strings.Contains(body, password) {
		t.Fatal("falta el error de longitud o se devolvió la contraseña")
	}
	userForm.Set("contrasena", password)
	post("/usuarios", userForm, http.StatusSeeOther)
	usuarios, err := queries.ListUsuarios(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range usuarios {
		if item.NombreUsuario == username {
			userID = item.IDUsuario
		}
	}
	if userID == 0 {
		t.Fatal("falta el usuario creado")
	}
	var stored string
	if err := conn.QueryRow("SELECT contrasena_hash FROM usuario WHERE id_usuario = $1", userID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stored, "pbkdf2-sha256$") || strings.Contains(stored, password) {
		t.Fatal("la contraseña no fue protegida")
	}
	page = post("/usuarios", userForm, http.StatusConflict)
	if !strings.Contains(page, username) || strings.Contains(page, password) {
		t.Fatal("error de duplicado no preserva campos seguros o devuelve la contraseña")
	}

	bookForm := url.Values{"titulo": {title}, "id_autor": {strconv.Itoa(int(authorID))}, "id_editorial": {strconv.Itoa(int(editorialID))}, "cantidad_paginas": {"250"}, "fecha_lanzamiento": {"2020-01-15"}}
	post("/libros", bookForm, http.StatusSeeOther)
	libros, err := queries.ListLibros(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range libros {
		if item.Titulo == title {
			bookID = item.IDLibro
		}
	}
	if bookID == 0 {
		t.Fatal("falta el libro creado")
	}
	page = get()
	if !strings.Contains(page, title) || !strings.Contains(page, authorName) || strings.Contains(page, password) {
		t.Fatal("listado incompleto o contraseña visible")
	}
	post(fmt.Sprintf("/autores/%d/eliminar", authorID), nil, http.StatusConflict)

	t.Run("validaciones del servidor", func(t *testing.T) {
		post("/editoriales", url.Values{"nombre": {"   "}, "sede": {"Tandil"}}, http.StatusBadRequest)
		bookForm.Set("cantidad_paginas", "2147483648")
		post("/libros", bookForm, http.StatusBadRequest)
		bookForm.Set("cantidad_paginas", "250")
		bookForm.Set("fecha_lanzamiento", "2099-01-15")
		post("/libros", bookForm, http.StatusBadRequest)
		bookForm.Set("fecha_lanzamiento", "2020-01-15")
		bookForm.Set("id_autor", "2147483647")
		post("/libros", bookForm, http.StatusConflict)
		post("/editoriales?nombre=NoDebeUsarse&sede=Tandil", nil, http.StatusBadRequest)
		request(http.MethodPost, "/editoriales", `{"nombre":"JSON"}`, "application/json", http.StatusBadRequest)
		post("/editoriales", url.Values{"nombre": {strings.Repeat("x", 70<<10)}}, http.StatusBadRequest)
	})
	t.Run("rutas y borrado clásico", func(t *testing.T) {
		request(http.MethodGet, fmt.Sprintf("/libros/%d/eliminar", bookID), "", "", http.StatusMethodNotAllowed)
		post("/libros/no-es-id/eliminar", nil, http.StatusBadRequest)
		post("/libros/2147483647/eliminar", nil, http.StatusNotFound)
		post("/desconocida", nil, http.StatusNotFound)
		request(http.MethodGet, "/no-existe", "", "", http.StatusNotFound)
		post(fmt.Sprintf("/libros/%d/eliminar", bookID), nil, http.StatusSeeOther)
		if _, err := queries.GetLibro(ctx, bookID); err != sql.ErrNoRows {
			t.Fatalf("el libro no se eliminó: %v", err)
		}
		bookID = 0
		post(fmt.Sprintf("/autores/%d/eliminar", authorID), nil, http.StatusSeeOther)
		authorID = 0
		post(fmt.Sprintf("/editoriales/%d/eliminar", editorialID), nil, http.StatusSeeOther)
		editorialID = 0
		post(fmt.Sprintf("/usuarios/%d/eliminar", userID), nil, http.StatusSeeOther)
		userID = 0
	})
}
