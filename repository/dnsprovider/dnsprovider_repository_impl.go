package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/dnsprovider"
	"autoSSL/ent/user"
)

type dnsProviderRepositoryImpl struct {
	client *ent.Client
}

func NewDNSProviderRepository(client *ent.Client) DNSProviderRepository {
	return &dnsProviderRepositoryImpl{client: client}
}

func (r *dnsProviderRepositoryImpl) Create(ctx context.Context, userID int, providerType string, config map[string]interface{}) (*ent.DNSProvider, error) {
	return r.client.DNSProvider.Create().
		SetProviderType(dnsprovider.ProviderType(providerType)).
		SetConfig(config).
		SetUserID(userID).
		Save(ctx)
}

func (r *dnsProviderRepositoryImpl) GetByID(ctx context.Context, userID, id int) (*ent.DNSProvider, error) {
	return r.client.DNSProvider.Query().
		Where(dnsprovider.ID(id), dnsprovider.HasUserWith(user.ID(userID))).
		Only(ctx)
}

func (r *dnsProviderRepositoryImpl) GetByIDForValidation(ctx context.Context, id int) (*ent.DNSProvider, error) {
	return r.client.DNSProvider.Query().
		Where(dnsprovider.ID(id)).
		Only(ctx)
}

func (r *dnsProviderRepositoryImpl) GetAll(ctx context.Context, userID int, page, pageSize int) ([]*ent.DNSProvider, int, error) {
	baseQuery := r.client.DNSProvider.Query().Where(dnsprovider.HasUserWith(user.ID(userID)))

	total, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	providers, err := baseQuery.
		Select(dnsprovider.FieldID, dnsprovider.FieldProviderType, dnsprovider.FieldCreatedAt, dnsprovider.FieldUpdatedAt).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return providers, total, nil
}

func (r *dnsProviderRepositoryImpl) Update(ctx context.Context, id int, providerType string, config map[string]interface{}) (*ent.DNSProvider, error) {
	return r.client.DNSProvider.UpdateOneID(id).
		SetProviderType(dnsprovider.ProviderType(providerType)).
		SetConfig(config).
		Save(ctx)
}

func (r *dnsProviderRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.client.DNSProvider.DeleteOneID(id).Exec(ctx)
}

func (r *dnsProviderRepositoryImpl) CountByUserID(ctx context.Context, userID int) (int, error) {
	return r.client.DNSProvider.Query().
		Where(dnsprovider.HasUserWith(user.ID(userID))).
		Count(ctx)
}
