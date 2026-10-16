package logic

import "context"

type EditorialRepository interface {
	Create(ctx context.Context, nombre, sede string) (*Editorial, error)
	FindById(ctx context.Context, id int32) (*Editorial, error)
	List(ctx context.Context) ([]*Editorial, error)
	Update(ctx context.Context, id int32, nombre, sede string) (*Editorial, error)
	Delete(ctx context.Context, id int32) error
}
