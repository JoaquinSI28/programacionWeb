package logic

import "context"

type UsuarioRepository interface {
	Create(ctx context.Context, nombre, email, contrasena string) (*Usuario, error)
	FindById(ctx context.Context, id int32) (*Usuario, error)
	List(ctx context.Context) ([]*Usuario, error)
	Delete(ctx context.Context, id int32) error
	Update(ctx context.Context, id int32, nombre, email, contrasena string) (*Usuario, error)
}
