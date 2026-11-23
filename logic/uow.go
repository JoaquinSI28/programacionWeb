package logic

import "context"

// RepositoryProvider expone repositorios vinculados a una misma unidad transaccional.
type RepositoryProvider interface {
	Autores() AutorRepository
	Libros() LibroRepository
}

// UnitOfWork ejecuta un conjunto de operaciones de forma atómica sin acoplar la lógica a SQL.
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(RepositoryProvider) error) error
}
