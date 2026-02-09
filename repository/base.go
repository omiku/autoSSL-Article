package repository

import "context"

type IRepository[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	GetByID(ctx context.Context, id int) (*T, error)
	Update(ctx context.Context, id int, entity *T) (*T, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*T, int, error)
	Count(ctx context.Context) (int, error)
}
