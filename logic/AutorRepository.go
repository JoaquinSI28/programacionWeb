package logic

import "context"

type AutorRepository interface {
	Create(ctx context.Context, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error)
	FindById(ctx context.Context, id int32) (*Autor, error)
	List(ctx context.Context) ([]*Autor, error)
	Update(ctx context.Context, id int32, nombre, pais string, fechaNacimiento Date, biografia string) (*Autor, error)
	Delete(ctx context.Context, id int32) error
}
