package logic_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bookfly.com/logic"
)

type mockAutorRepo struct {
	autores       map[int32]*logic.Autor
	updateLlamado bool
}

func (m *mockAutorRepo) Create(_ context.Context, nombre, pais string, fechaNacimiento logic.Date, biografia string) (*logic.Autor, error) {
	id := int32(len(m.autores) + 1)
	a := &logic.Autor{IDAutor: id, Nombre: nombre, Pais: pais, FechaNacimiento: fechaNacimiento, Biografia: biografia}
	m.autores[id] = a
	return a, nil
}

func (m *mockAutorRepo) FindById(_ context.Context, id int32) (*logic.Autor, error) {
	a, ok := m.autores[id]
	if !ok {
		return nil, logic.ErrAutorNoEncontrado
	}
	return a, nil
}

func (m *mockAutorRepo) List(_ context.Context) ([]*logic.Autor, error) {
	var res []*logic.Autor
	for _, a := range m.autores {
		res = append(res, a)
	}
	return res, nil
}

func (m *mockAutorRepo) Update(_ context.Context, id int32, nombre, pais string, fechaNacimiento logic.Date, biografia string) (*logic.Autor, error) {
	m.updateLlamado = true
	if _, ok := m.autores[id]; !ok {
		return nil, logic.ErrAutorNoEncontrado
	}
	a := &logic.Autor{IDAutor: id, Nombre: nombre, Pais: pais, FechaNacimiento: fechaNacimiento, Biografia: biografia}
	m.autores[id] = a
	return a, nil
}

func (m *mockAutorRepo) Delete(_ context.Context, id int32) error {
	if _, ok := m.autores[id]; !ok {
		return logic.ErrAutorNoEncontrado
	}
	delete(m.autores, id)
	return nil
}

type mockLibroRepo struct {
	libros        map[int32]*logic.Libro
	createLlamado bool
	updateLlamado bool
}

func (m *mockLibroRepo) Create(_ context.Context, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento logic.Date, cantidadPaginas int32) (*logic.Libro, error) {
	m.createLlamado = true
	id := int32(len(m.libros) + 1)
	l := &logic.Libro{IDLibro: id, Titulo: titulo, IDAutor: idAutor, IDEditorial: idEditorial, FechaLanzamiento: fechaLanzamiento, CantidadPaginas: cantidadPaginas}
	m.libros[id] = l
	return l, nil
}

func (m *mockLibroRepo) FindById(_ context.Context, id int32) (*logic.Libro, error) {
	l, ok := m.libros[id]
	if !ok {
		return nil, logic.ErrLibroNoEncontrado
	}
	return l, nil
}

func (m *mockLibroRepo) List(_ context.Context) ([]*logic.Libro, error) {
	var res []*logic.Libro
	for _, l := range m.libros {
		res = append(res, l)
	}
	return res, nil
}

func (m *mockLibroRepo) Delete(_ context.Context, id int32) error {
	if _, ok := m.libros[id]; !ok {
		return logic.ErrLibroNoEncontrado
	}
	delete(m.libros, id)
	return nil
}

func (m *mockLibroRepo) Update(_ context.Context, id int32, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento logic.Date, cantidadPaginas int32) (*logic.Libro, error) {
	m.updateLlamado = true
	if _, ok := m.libros[id]; !ok {
		return nil, logic.ErrLibroNoEncontrado
	}
	l := &logic.Libro{IDLibro: id, Titulo: titulo, IDAutor: idAutor, IDEditorial: idEditorial, FechaLanzamiento: fechaLanzamiento, CantidadPaginas: cantidadPaginas}
	m.libros[id] = l
	return l, nil
}

func (m *mockLibroRepo) PrimerLanzamientoDelAutor(_ context.Context, idAutor int32) (logic.Date, bool, error) {
	var primero logic.Date
	encontrado := false
	for _, l := range m.libros {
		if l.IDAutor == idAutor && (!encontrado || l.FechaLanzamiento.Before(primero)) {
			primero = l.FechaLanzamiento
			encontrado = true
		}
	}
	return primero, encontrado, nil
}

type mockUnitOfWork struct {
	autores *mockAutorRepo
	libros  *mockLibroRepo
}

func (u *mockUnitOfWork) Autores() logic.AutorRepository { return u.autores }
func (u *mockUnitOfWork) Libros() logic.LibroRepository  { return u.libros }
func (u *mockUnitOfWork) Execute(_ context.Context, fn func(logic.RepositoryProvider) error) error {
	return fn(u)
}

func newMockUoW() *mockUnitOfWork {
	return &mockUnitOfWork{
		autores: &mockAutorRepo{autores: make(map[int32]*logic.Autor)},
		libros:  &mockLibroRepo{libros: make(map[int32]*logic.Libro)},
	}
}

func TestReglaEdadMinimaEnLogicConMocks(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateLibro rechaza antes del quinto cumpleaños sin invocar persistencia", func(t *testing.T) {
		uow := newMockUoW()
		uow.autores.autores[1] = &logic.Autor{
			IDAutor: 1, Nombre: "Ana", Pais: "Argentina",
			FechaNacimiento: time.Date(1980, 4, 12, 0, 0, 0, 0, time.UTC), Biografia: "Bio",
		}
		_, err := logic.CreateLibro(ctx, uow, 1, "Primer libro", 1, time.Date(1985, 4, 11, 0, 0, 0, 0, time.UTC), 100)
		var validacion *logic.ErrorValidacion
		if !errors.As(err, &validacion) {
			t.Fatalf("se esperaba error de validación, se obtuvo %v", err)
		}
		if uow.libros.createLlamado {
			t.Fatal("no debía persistirse un libro que incumple la edad mínima")
		}
	})

	t.Run("CreateLibro acepta en el quinto cumpleaños y en 29 de febrero", func(t *testing.T) {
		uow := newMockUoW()
		uow.autores.autores[1] = &logic.Autor{
			IDAutor: 1, Nombre: "Ana", Pais: "Argentina",
			FechaNacimiento: time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC), Biografia: "Bio",
		}
		if _, err := logic.CreateLibro(ctx, uow, 1, "Bisiesto temprano", 1, time.Date(2005, 2, 27, 0, 0, 0, 0, time.UTC), 50); err == nil {
			t.Fatal("debió rechazar el día previo al 28 de febrero")
		}
		creado, err := logic.CreateLibro(ctx, uow, 1, "Bisiesto válido", 1, time.Date(2005, 2, 28, 0, 0, 0, 0, time.UTC), 50)
		if err != nil {
			t.Fatalf("debió aceptar el 28 de febrero como quinto aniversario: %v", err)
		}
		if creado.Titulo != "Bisiesto válido" {
			t.Fatalf("título inesperado: %q", creado.Titulo)
		}
	})

	t.Run("CreateLibro traduce autor inexistente a ErrReferencia", func(t *testing.T) {
		uow := newMockUoW()
		_, err := logic.CreateLibro(ctx, uow, 99, "Sin autor", 1, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), 50)
		if !errors.Is(err, logic.ErrReferencia) {
			t.Fatalf("se esperaba ErrReferencia, se obtuvo %v", err)
		}
	})

	t.Run("UpdateLibro rechaza lanzamiento inválido sin actualizar", func(t *testing.T) {
		uow := newMockUoW()
		uow.autores.autores[1] = &logic.Autor{
			IDAutor: 1, Nombre: "Ana", Pais: "Argentina",
			FechaNacimiento: time.Date(1980, 4, 12, 0, 0, 0, 0, time.UTC), Biografia: "Bio",
		}
		uow.libros.libros[1] = &logic.Libro{
			IDLibro: 1, Titulo: "Original", IDAutor: 1, IDEditorial: 1,
			FechaLanzamiento: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), CantidadPaginas: 100,
		}
		_, err := logic.UpdateLibro(ctx, uow, 1, 1, "Original", 1, time.Date(1984, 12, 31, 0, 0, 0, 0, time.UTC), 100)
		var validacion *logic.ErrorValidacion
		if !errors.As(err, &validacion) {
			t.Fatalf("se esperaba error de validación, se obtuvo %v", err)
		}
		if uow.libros.updateLlamado {
			t.Fatal("no debía actualizarse un libro con fecha inválida")
		}
	})

	t.Run("UpdateAutor valida contra el primer lanzamiento del autor", func(t *testing.T) {
		uow := newMockUoW()
		uow.autores.autores[1] = &logic.Autor{
			IDAutor: 1, Nombre: "Ana", Pais: "Argentina",
			FechaNacimiento: time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC), Biografia: "Bio",
		}
		uow.libros.libros[1] = &logic.Libro{
			IDLibro: 1, Titulo: "Cuentos", IDAutor: 1, IDEditorial: 1,
			FechaLanzamiento: time.Date(2005, 2, 28, 0, 0, 0, 0, time.UTC), CantidadPaginas: 20,
		}
		_, err := logic.UpdateAutor(ctx, uow, 1, "Ana", "Argentina", time.Date(2000, 3, 1, 0, 0, 0, 0, time.UTC), "Bio")
		var validacion *logic.ErrorValidacion
		if !errors.As(err, &validacion) {
			t.Fatalf("se esperaba rechazo por edad al mover nacimiento, se obtuvo %v", err)
		}
		if uow.autores.updateLlamado {
			t.Fatal("no debía actualizarse el autor si invalida libros existentes")
		}

		actualizado, err := logic.UpdateAutor(ctx, uow, 1, "Ana María", "Argentina", time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC), "Bio")
		if err != nil {
			t.Fatalf("debió permitir adelantar el nacimiento: %v", err)
		}
		if actualizado.Nombre != "Ana María" {
			t.Fatalf("nombre inesperado: %q", actualizado.Nombre)
		}
	})

	t.Run("UpdateAutor permite cambiar nacimiento si el autor no tiene libros", func(t *testing.T) {
		uow := newMockUoW()
		uow.autores.autores[1] = &logic.Autor{
			IDAutor: 1, Nombre: "Ana", Pais: "Argentina",
			FechaNacimiento: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC), Biografia: "Bio",
		}
		if _, err := logic.UpdateAutor(ctx, uow, 1, "Ana", "Argentina", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC), "Bio"); err != nil {
			t.Fatalf("no debía fallar si el autor no tiene libros: %v", err)
		}
	})
}
