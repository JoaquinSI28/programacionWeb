package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"bookfly.com/db/migrations"
	db "bookfly.com/db/sqlc"
	"bookfly.com/logic"
	"bookfly.com/repository"
	"github.com/lib/pq"
)

// Cada conexión usa el mismo esquema de prueba, aislado del catálogo y de la suite HTTP.
func catalogoAnterior(t *testing.T) *sql.DB {
	t.Helper()
	esquema := fmt.Sprintf("catalogo_prueba_%d", time.Now().UnixNano())
	conn, err := sql.Open("postgres", testDSN()+"&search_path="+esquema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if _, err := conn.Exec("CREATE SCHEMA " + esquema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := conn.Exec("DROP SCHEMA " + esquema + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	schema, err := os.ReadFile("schema/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestMigracionCatalogoConDuplicados(t *testing.T) {
	conn := catalogoAnterior(t)
	ctx := context.Background()
	if _, err := conn.Exec("INSERT INTO editorial (nombre, sede) VALUES ('Sur', 'Tandil'), ('  SUR', 'Tandil')"); err != nil {
		t.Fatal(err)
	}
	err := migrations.AplicarReglas(ctx, conn)
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("se esperaba un conflicto al crear el índice, se obtuvo %v", err)
	}
	var cantidad int
	if err := conn.QueryRow("SELECT count(*) FROM editorial").Scan(&cantidad); err != nil || cantidad != 2 {
		t.Fatalf("el catálogo debe conservar los dos registros: cantidad=%d, error=%v", cantidad, err)
	}
	if _, err := conn.Exec("UPDATE editorial SET sede = 'Azul' WHERE nombre = '  SUR'"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.AplicarReglas(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := migrations.AplicarReglas(ctx, conn); err != nil {
		t.Fatalf("falló el segundo arranque: %v", err)
	}
}

// Las escrituras son incompatibles: una sola puede quedar confirmada.
func conflictoConcurrente(t *testing.T, primera, segunda func() error) error {
	t.Helper()
	inicio := make(chan struct{})
	resultados := make(chan error, 2)
	for _, escribir := range []func() error{primera, segunda} {
		go func() {
			<-inicio
			resultados <- escribir()
		}()
	}
	close(inicio)
	a, b := <-resultados, <-resultados
	if (a == nil) == (b == nil) {
		t.Fatalf("se esperaba un éxito y un rechazo: %v, %v", a, b)
	}
	return errors.Join(a, b)
}

func TestReglasConEscriturasConcurrentes(t *testing.T) {
	conn := catalogoAnterior(t)
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()
	if err := migrations.AplicarReglas(ctx, conn); err != nil {
		t.Fatal(err)
	}
	queries := db.New(conn)
	editoriales := repository.NewEditorialSQL(queries)
	t.Run("dos altas de la misma editorial", func(t *testing.T) {
		err := conflictoConcurrente(t, func() error {
			_, err := editoriales.Create(ctx, "Estrella del Sur", "Tandil")
			return err
		}, func() error {
			_, err := editoriales.Create(ctx, "  ESTRELLA   del Sur", "TANDIL ")
			return err
		})
		if !errors.Is(err, logic.ErrDuplicado) {
			t.Fatalf("se esperaba un duplicado, se obtuvo %v", err)
		}
	})
	t.Run("dos bajas de la misma editorial", func(t *testing.T) {
		editorial, err := editoriales.Create(ctx, "Luna", "Olavarría")
		if err != nil {
			t.Fatal(err)
		}
		borrar := func() error { return editoriales.Delete(ctx, editorial.IDEditorial) }
		if err := conflictoConcurrente(t, borrar, borrar); !errors.Is(err, logic.ErrEditorialNoEncontrada) {
			t.Fatalf("la segunda baja debe informar que la editorial ya no existe: %v", err)
		}
		if _, err := editoriales.Update(ctx, editorial.IDEditorial, editorial.Nombre, "Azul"); !errors.Is(err, logic.ErrEditorialNoEncontrada) {
			t.Fatalf("la edición posterior a la baja debe informar que la editorial ya no existe: %v", err)
		}
	})
	t.Run("publicación y cambio de nacimiento", func(t *testing.T) {
		autores := repository.NewAutorSQL(conn)
		libros := repository.NewLibroSQL(conn)
		editorial, err := editoriales.Create(ctx, "La Página", "Azul")
		if err != nil {
			t.Fatal(err)
		}
		nacimiento := time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC)
		autor, err := autores.Create(ctx, "Ana", "Argentina", nacimiento, "Biografía")
		if err != nil {
			t.Fatal(err)
		}
		lanzamiento := time.Date(2005, 2, 28, 0, 0, 0, 0, time.UTC)
		err = conflictoConcurrente(t, func() error {
			_, err := libros.Create(ctx, autor.IDAutor, "Cuentos", editorial.IDEditorial, lanzamiento, 20)
			return err
		}, func() error {
			_, err := autores.Update(ctx, autor.IDAutor, autor.Nombre, autor.Pais, nacimiento.AddDate(0, 0, 1), autor.Biografia)
			return err
		})
		var validacion *logic.ErrorValidacion
		if !errors.As(err, &validacion) {
			t.Fatalf("se esperaba un rechazo por edad, se obtuvo %v", err)
		}
	})
}
