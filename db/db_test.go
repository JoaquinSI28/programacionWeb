package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	db "bookfly.com/db/sqlc"
	_ "github.com/lib/pq"
)

func newQueries(t *testing.T) *db.Queries {
	t.Helper()

	dsn := fmt.Sprintf("postgresql://%s:%s@localhost:5432/%s?sslmode=disable",
		envOrDefault("POSTGRES_USER", "postgres"),
		envOrDefault("POSTGRES_PASSWORD", "postgres"),
		envOrDefault("POSTGRES_DB", "prueba"))
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("No se pudo conectar: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return db.New(conn)
}

func TestUsuario_CRUD(t *testing.T) {
	ctx := context.Background()
	queries := newQueries(t)

	// 1. CreateUsuario
	createParams := db.CreateUsuarioParams{
		NombreUsuario: "usuario_test",
		Mail:          "test@example.com",
		Contrasena:    "secreta123",
	}
	user, err := queries.CreateUsuario(ctx, createParams)
	if err != nil {
		t.Fatalf("Fallo en CreateUsuario: %v", err)
	}
	if user.NombreUsuario != createParams.NombreUsuario {
		t.Errorf("Se esperaba %s, se obtuvo %s", createParams.NombreUsuario, user.NombreUsuario)
	}
	t.Logf("Usuario creado exitosamente con ID: %d", user.IDUsuario)

	// 2. GetUsuario
	_, err = queries.GetUsuario(ctx, user.IDUsuario)
	if err != nil {
		t.Fatalf("Fallo en GetUsuario: %v", err)
	}
	t.Log("Usuario obtenido exitosamente")

	// 3. UpdateUsuario
	updateParams := db.UpdateUsuarioParams{
		IDUsuario:     user.IDUsuario,
		NombreUsuario: "usuario_modificado",
		Mail:          "mod@example.com",
		Contrasena:    "nueva_secreta",
	}
	err = queries.UpdateUsuario(ctx, updateParams)
	if err != nil {
		t.Fatalf("Fallo en UpdateUsuario: %v", err)
	}
	t.Log("Usuario actualizado exitosamente")

	// 4. ListUsuarios
	users, err := queries.ListUsuarios(ctx)
	if err != nil {
		t.Fatalf("Fallo en ListUsuarios: %v", err)
	}
	encontrado := false
	for _, u := range users {
		if u.IDUsuario == user.IDUsuario && u.NombreUsuario == updateParams.NombreUsuario {
			encontrado = true
			break
		}
	}
	if !encontrado {
		t.Error("El usuario actualizado no se encontró en la lista")
	}
	t.Log("Usuario encontrado y validado en la lista exitosamente")

	// 5. DeleteUsuario
	err = queries.DeleteUsuario(ctx, user.IDUsuario)
	if err != nil {
		t.Fatalf("Fallo en DeleteUsuario: %v", err)
	}
	t.Log("Usuario eliminado exitosamente")

	_, err = queries.GetUsuario(ctx, user.IDUsuario)
	if err != sql.ErrNoRows {
		t.Errorf("Se esperaba sql.ErrNoRows, se obtuvo: %v", err)
	}
	t.Log("Verificación de borrado exitosa (el usuario ya no existe)")
}

func TestLibro_CRUD(t *testing.T) {
	ctx := context.Background()
	queries := newQueries(t)
	fechaLanzamiento := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	fechaNacimiento := time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

	autor, err := queries.CreateAutor(ctx, db.CreateAutorParams{
		Nombre:          "Autor libro test",
		Pais:            "Argentina",
		FechaNacimiento: fechaNacimiento,
		Biografia:       "Biografía de prueba",
	})
	if err != nil {
		t.Fatalf("Fallo en CreateAutor: %v", err)
	}
	defer queries.DeleteAutor(ctx, autor.IDAutor)

	editorial, err := queries.CreateEditorial(ctx, db.CreateEditorialParams{
		Nombre: "Editorial libro test",
		Sede:   "Buenos Aires",
	})
	if err != nil {
		t.Fatalf("Fallo en CreateEditorial: %v", err)
	}
	defer queries.DeleteEditorial(ctx, editorial.IDEditorial)

	createParams := db.CreateLibroParams{
		Titulo:           "Libro de prueba",
		IDAutor:          autor.IDAutor,
		IDEditorial:      editorial.IDEditorial,
		FechaLanzamiento: fechaLanzamiento,
		CantidadPaginas:  250,
	}
	libro, err := queries.CreateLibro(ctx, createParams)
	if err != nil {
		t.Fatalf("Fallo en CreateLibro: %v", err)
	}
	t.Logf("Libro creado exitosamente con ID: %d", libro.IDLibro)

	defer queries.DeleteLibro(ctx, libro.IDLibro)

	libroObtenido, err := queries.GetLibro(ctx, libro.IDLibro)
	if err != nil {
		t.Fatalf("Fallo en GetLibro: %v", err)
	}
	if libroObtenido.Titulo != createParams.Titulo {
		t.Errorf("Se esperaba %s, se obtuvo %s", createParams.Titulo, libroObtenido.Titulo)
	}
	t.Log("Libro obtenido exitosamente")

	updateParams := db.UpdateLibroParams{
		IDLibro:          libro.IDLibro,
		Titulo:           "Libro de prueba actualizado",
		IDAutor:          createParams.IDAutor,
		IDEditorial:      createParams.IDEditorial,
		FechaLanzamiento: createParams.FechaLanzamiento,
		CantidadPaginas:  300,
	}
	if err = queries.UpdateLibro(ctx, updateParams); err != nil {
		t.Fatalf("Fallo en UpdateLibro: %v", err)
	}
	t.Log("Libro actualizado exitosamente")

	libros, err := queries.ListLibros(ctx)
	if err != nil {
		t.Fatalf("Fallo en ListLibros: %v", err)
	}
	encontrado := false
	for _, item := range libros {
		if item.IDLibro == libro.IDLibro && item.Titulo == updateParams.Titulo {
			encontrado = true
			break
		}
	}
	if !encontrado {
		t.Error("El libro actualizado no se encontró en la lista")
	}
	t.Log("Libro encontrado y validado en la lista exitosamente")

	if err = queries.DeleteLibro(ctx, libro.IDLibro); err != nil {
		t.Fatalf("Fallo en DeleteLibro: %v", err)
	}
	t.Log("Libro eliminado exitosamente")

	_, err = queries.GetLibro(ctx, libro.IDLibro)
	if err != sql.ErrNoRows {
		t.Errorf("Se esperaba sql.ErrNoRows, se obtuvo: %v", err)
	}
	t.Log("Verificación de borrado del libro exitosa")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
