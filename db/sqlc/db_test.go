package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestUsuario_CRUD(t *testing.T) {
	ctx := context.Background()

	conn, err := sql.Open("postgres", "postgresql://postgres:postgres@localhost:5432/prueba?sslmode=disable")
	if err != nil {
		t.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()

	queries := New(conn)
	t.Log("Conexión a la base de datos inicializada exitosamente")

	// 1. CreateUsuario
	createParams := CreateUsuarioParams{ 
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
	updateParams := UpdateUsuarioParams{ 
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
