package repository

import (
	"context"

	"autoSSL/ent"
)

type AccessRepository interface {
	Create(ctx context.Context, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error)
	GetByID(ctx context.Context, userID, id int) (*ent.Access, error)
	ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Access, error)
	Update(ctx context.Context, id int, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error)
	Delete(ctx context.Context, id int, userID int) error
	GetByIDs(ctx context.Context, ids []int) ([]*ent.Access, error)
	GetByServerGroupID(ctx context.Context, serverGroupID int) ([]*ent.Access, error)
	CountByUserID(ctx context.Context, userID int) (int, error)
}
