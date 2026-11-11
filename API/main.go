package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bookfly.com/db/migrations"
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

	if err := repository.MigrarContrasenas(context.Background(), dbConn); err != nil {
		log.Fatalf("Error al preparar usuarios: %v", err)
	}
	if err := migrations.AplicarReglas(context.Background(), dbConn); err != nil {
		log.Fatalf("Error al aplicar las reglas del catálogo: %v", err)
	}
	queries := dbsqlc.New(dbConn)

	port := ":8080"
	log.Printf("Servidor escuchando en http://localhost%s", port)
	server := &http.Server{
		Addr: port, Handler: newWebHandler(queries),
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
