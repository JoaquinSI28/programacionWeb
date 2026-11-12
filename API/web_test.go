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

	"bookfly.com/db/migrations"
	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/logic"
)

// Se ejecuta con make test sobre un PostgreSQL de prueba preparado por Compose.
func TestHTMLFragments(t *testing.T) {
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
	if err := migrations.AplicarReglas(context.Background(), conn); err != nil {
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
		if method == http.MethodDelete && want == http.StatusConflict {
			parts := strings.Split(path, "/")
			expected := "#" + parts[1] + "-error-" + parts[2]
			if response.Header.Get("HX-Retarget") != expected || !strings.Contains(string(data), "tiene libros asociados") {
				t.Fatal("el conflicto de borrado no indica el mensaje y su destino junto al registro")
			}
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: estado %d, esperado %d; %s", method, path, response.StatusCode, want, data)
		}
		if want == http.StatusOK && response.Header.Get("Location") != "" {
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
	if !strings.Contains(page, "htmx.org@2.0.11") || !strings.Contains(page, "hx-post") {
		t.Fatal("falta la integración HTMX")
	}
	request(http.MethodGet, "/static/styles.css", "", "", http.StatusOK)
	request(http.MethodGet, "/app.js", "", "", http.StatusNotFound)
	fragment := post("/editoriales", url.Values{"nombre": {name}, "sede": {"Tandil"}}, http.StatusOK)
	if strings.Contains(fragment, "<html") || !strings.Contains(fragment, "editoriales-list") {
		t.Fatal("el alta no devolvió la lista parcial")
	}
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

	post("/autores", url.Values{"nombre": {authorName}, "pais": {"Argentina"}, "fecha_nacimiento": {"1980-01-15"}, "biografia": {"Biografía"}}, http.StatusOK)
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
	post("/usuarios", userForm, http.StatusOK)
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
	if !strings.Contains(page, "ya está registrado") || strings.Contains(page, password) {
		t.Fatal("error de duplicado no preserva campos seguros o devuelve la contraseña")
	}

	bookForm := url.Values{"titulo": {title}, "id_autor": {strconv.Itoa(int(authorID))}, "id_editorial": {strconv.Itoa(int(editorialID))}, "cantidad_paginas": {"250"}, "fecha_lanzamiento": {"2020-01-15"}}
	post("/libros", bookForm, http.StatusOK)
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
	request(http.MethodDelete, fmt.Sprintf("/autores/%d", authorID), "", "", http.StatusConflict)

	t.Run("duplicados y fechas del catálogo", func(t *testing.T) {
		autor := url.Values{"nombre": {strings.ToUpper(authorName)}, "pais": {"  argentina  "}, "fecha_nacimiento": {"1980-01-15"}, "biografia": {"Otra biografía"}}
		if body := post("/autores", autor, http.StatusConflict); !strings.Contains(body, "Ya existe un autor") {
			t.Fatal("mensaje de autor duplicado incorrecto")
		}
		hoy, err := time.Parse(time.DateOnly, logic.Hoy())
		if err != nil {
			t.Fatal(err)
		}
		for _, fecha := range []string{"", "0000-01-01", "2025-02-29", hoy.AddDate(0, 0, 1).Format(time.DateOnly)} {
			autor.Set("fecha_nacimiento", fecha)
			post("/autores", autor, http.StatusBadRequest)
		}
		defer conn.Exec("DELETE FROM autor WHERE nombre = $1 AND fecha_nacimiento = $2", autor.Get("nombre"), hoy)
		autor.Set("fecha_nacimiento", hoy.Format(time.DateOnly))
		post("/autores", autor, http.StatusOK)
		editorial := url.Values{"nombre": {strings.ReplaceAll(strings.ToUpper(name), " ", "  ")}, "sede": {"tandil"}}
		post("/editoriales", editorial, http.StatusConflict)
		defer conn.Exec("DELETE FROM editorial WHERE nombre = $1 AND sede = $2", editorial.Get("nombre"), "Otra sede")
		editorial.Set("sede", "Otra sede")
		post("/editoriales", editorial, http.StatusOK)
		bookForm.Set("cantidad_paginas", "251")
		post("/libros", bookForm, http.StatusConflict)
		bookForm.Set("fecha_lanzamiento", "2020-02-29")
		post("/libros", bookForm, http.StatusOK)
		bookForm.Set("fecha_lanzamiento", "2020-01-15")
		post("/usuarios", url.Values{"nombre": {strings.ToUpper(username)}, "email": {"otro" + username + "@example.com"}, "contrasena": {"clave"}}, http.StatusConflict)
		post("/usuarios", url.Values{"nombre": {"otro" + username}, "email": {strings.ToUpper(userForm.Get("email"))}, "contrasena": {"clave"}}, http.StatusConflict)
		post("/editoriales", url.Values{"nombre": {strings.Repeat("ñ", 256)}, "sede": {"Tandil"}}, http.StatusBadRequest)
		if _, err := conn.Exec("UPDATE autor SET fecha_nacimiento = DATE '9999-01-01' WHERE id_autor = $1", authorID); err == nil {
			t.Fatal("la base aceptó un nacimiento futuro")
		}
		if _, err := conn.Exec("DELETE FROM libro WHERE titulo = $1 AND fecha_lanzamiento = DATE '2020-02-29'", title); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("validaciones del servidor", func(t *testing.T) {
		post("/editoriales", url.Values{"nombre": {"   "}, "sede": {"Tandil"}}, http.StatusBadRequest)
		bookForm.Set("fecha_lanzamiento", "1979-12-31")
		if body := post("/libros", bookForm, http.StatusBadRequest); !strings.Contains(body, "5 años cumplidos") {
			t.Fatal("falta el motivo de la cronología inválida")
		}
		if _, err := conn.Exec("UPDATE libro SET fecha_lanzamiento = DATE '1979-12-31' WHERE id_libro = $1", bookID); err == nil {
			t.Fatal("la base permitió un libro anterior al nacimiento")
		}
		if _, err := conn.Exec("UPDATE autor SET fecha_nacimiento = DATE '2021-01-01' WHERE id_autor = $1", authorID); err == nil {
			t.Fatal("la base permitió invalidar libros al cambiar el autor")
		}
		// Se rechazan el nacimiento y el día anterior al quinto cumpleaños.
		bookForm.Set("fecha_lanzamiento", "1980-01-15")
		post("/libros", bookForm, http.StatusBadRequest)
		bookForm.Set("fecha_lanzamiento", "1985-01-14")
		post("/libros", bookForm, http.StatusBadRequest)
		bookForm.Set("fecha_lanzamiento", "1985-01-15")
		post("/libros", bookForm, http.StatusOK)
		if _, err := conn.Exec("DELETE FROM libro WHERE titulo = $1 AND fecha_lanzamiento = DATE '1985-01-15'", title); err != nil {
			t.Fatal(err)
		}
		bookForm.Set("fecha_lanzamiento", "2020-01-15")
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
	t.Run("rutas y borrado HTMX", func(t *testing.T) {
		request(http.MethodGet, fmt.Sprintf("/libros/%d", bookID), "", "", http.StatusMethodNotAllowed)
		request(http.MethodDelete, "/libros/no-es-id", "", "", http.StatusBadRequest)
		request(http.MethodDelete, "/libros/2147483647", "", "", http.StatusNotFound)
		post("/desconocida", nil, http.StatusNotFound)
		request(http.MethodGet, "/no-existe", "", "", http.StatusNotFound)
		request(http.MethodDelete, fmt.Sprintf("/libros/%d", bookID), "", "", http.StatusOK)
		if _, err := queries.GetLibro(ctx, bookID); err != sql.ErrNoRows {
			t.Fatalf("el libro no se eliminó: %v", err)
		}
		bookID = 0
		request(http.MethodDelete, fmt.Sprintf("/autores/%d", authorID), "", "", http.StatusOK)
		authorID = 0
		request(http.MethodDelete, fmt.Sprintf("/editoriales/%d", editorialID), "", "", http.StatusOK)
		editorialID = 0
		request(http.MethodDelete, fmt.Sprintf("/usuarios/%d", userID), "", "", http.StatusOK)
		userID = 0
	})
}
