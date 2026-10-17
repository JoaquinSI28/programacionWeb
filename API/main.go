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
	http.Handle("/usuarios", usuarios)
	http.Handle("/usuarios/{id}", usuarios)
	http.Handle("/libros", libros)
	http.Handle("/libros/{id}", libros)
	http.Handle("/autores", autores)
	http.Handle("/autores/{id}", autores)
	http.Handle("/editoriales", editoriales)
	http.Handle("/editoriales/{id}", editoriales)
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
