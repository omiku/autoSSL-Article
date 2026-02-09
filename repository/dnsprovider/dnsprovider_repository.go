package repository

import (
	"context"

	"autoSSL/ent"
)

type DNSProviderRepository interface {
	Create(ctx context.Context, userID int, providerType string, config map[string]interface{}) (*ent.DNSProvider, error)
	GetByID(ctx context.Context, userID, id int) (*ent.DNSProvider, error)
	GetByIDForValidation(ctx context.Context, id int) (*ent.DNSProvider, error)
	GetAll(ctx context.Context, userID int, page, pageSize int) ([]*ent.DNSProvider, int, error)
	Update(ctx context.Context, id int, providerType string, config map[string]interface{}) (*ent.DNSProvider, error)
	Delete(ctx context.Context, id int) error
	CountByUserID(ctx context.Context, userID int) (int, error)
}
