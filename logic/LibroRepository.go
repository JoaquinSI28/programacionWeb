package logic

import "context"

type LibroRepository interface {
	Create(ctx context.Context, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error)
	FindById(ctx context.Context, id int32) (*Libro, error)
	List(ctx context.Context) ([]*Libro, error)
	Delete(ctx context.Context, id int32) error
	Update(ctx context.Context, id int32, idAutor int32, titulo string, idEditorial int32, fechaLanzamiento Date, cantidadPaginas int32) (*Libro, error)
}
