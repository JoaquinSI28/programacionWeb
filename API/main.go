package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	dbsqlc "bookfly.com/db/sqlc"
	"bookfly.com/repository"
	_ "github.com/lib/pq"
)

func main() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		envOrDefault("DB_USER", "postgres"),
		envOrDefault("DB_PASSWORD", "postgres"),
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_NAME", "prueba"),
	)

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error al configurar la base de datos: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v", err)
	}
	log.Println("Conexión a la base de datos exitosa")

	queries := dbsqlc.New(dbConn)
	usuarioRepo := repository.NewUsuarioSQL(queries)
	libroRepo := repository.NewLibroSQL(queries)
	autorRepo := repository.NewAutorSQL(queries)
	editorialRepo := repository.NewEditorialSQL(queries)
	usuarios := usuarioHandler(usuarioRepo)
	libros := libroHandler(libroRepo)
	autores := autorHandler(autorRepo)
	editoriales := editorialHandler(editorialRepo)
	http.Handle("GET /usuarios/{id}", usuarios)
	http.Handle("POST /usuarios", usuarios)
	http.Handle("GET /usuarios", usuarios)
	http.Handle("PUT /usuarios/{id}", usuarios)
	http.Handle("DELETE /usuarios/{id}", usuarios)
	http.Handle("GET /libros/{id}", libros)
	http.Handle("GET /libros", libros)
	http.Handle("POST /libros", libros)
	http.Handle("PUT /libros/{id}", libros)
	http.Handle("DELETE /libros/{id}", libros)
	http.Handle("GET /autores/{id}", autores)
	http.Handle("GET /autores", autores)
	http.Handle("POST /autores", autores)
	http.Handle("PUT /autores/{id}", autores)
	http.Handle("DELETE /autores/{id}", autores)
	http.Handle("GET /editoriales/{id}", editoriales)
	http.Handle("GET /editoriales", editoriales)
	http.Handle("POST /editoriales", editoriales)
	http.Handle("PUT /editoriales/{id}", editoriales)
	http.Handle("DELETE /editoriales/{id}", editoriales)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	port := ":8080"
	log.Printf("Servidor escuchando en http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
