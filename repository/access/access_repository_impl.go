package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/access"
	"autoSSL/ent/servergroup"
	"autoSSL/ent/servergroupmember"
	"autoSSL/ent/user"
)

type accessRepositoryImpl struct {
	client *ent.Client
}

func NewAccessRepository(client *ent.Client) AccessRepository {
	return &accessRepositoryImpl{client: client}
}

func (r *accessRepositoryImpl) Create(ctx context.Context, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
	return r.client.Access.Create().
		SetName(name).
		SetConfig(config).
		SetType(access.Type(accessType)).
		SetIsActive(isActive).
		SetUserID(userID).
		Save(ctx)
}

func (r *accessRepositoryImpl) GetByID(ctx context.Context, userID, id int) (*ent.Access, error) {
	return r.client.Access.Query().
		Where(access.ID(id), access.HasUserWith(user.ID(userID))).
		WithUser().
		WithServerGroupMembers(func(q *ent.ServerGroupMemberQuery) {
			q.WithServerGroup()
		}).
		Only(ctx)
}

func (r *accessRepositoryImpl) ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Access, error) {
	query := r.client.Access.Query().
		Where(access.HasUserWith(user.ID(userID))).
		WithUser()

	if page > 0 && pageSize > 0 {
		query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	return query.All(ctx)
}

func (r *accessRepositoryImpl) Update(ctx context.Context, id int, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
	return r.client.Access.UpdateOneID(id).
		Where(access.HasUserWith(user.ID(userID))).
		SetName(name).
		SetConfig(config).
		SetType(access.Type(accessType)).
		SetIsActive(isActive).
		Save(ctx)
}

func (r *accessRepositoryImpl) Delete(ctx context.Context, id int, userID int) error {
	return r.client.Access.DeleteOneID(id).
		Where(access.HasUserWith(user.ID(userID))).
		Exec(ctx)
}

func (r *accessRepositoryImpl) GetByIDs(ctx context.Context, ids []int) ([]*ent.Access, error) {
	if len(ids) == 0 {
		return []*ent.Access{}, nil
	}
	return r.client.Access.Query().
		Where(access.IDIn(ids...)).
		All(ctx)
}

func (r *accessRepositoryImpl) GetByServerGroupID(ctx context.Context, serverGroupID int) ([]*ent.Access, error) {
	return r.client.Access.Query().
		Where(
			access.HasServerGroupMembersWith(
				servergroupmember.HasServerGroupWith(
					servergroup.ID(serverGroupID),
				),
			),
		).
		All(ctx)
}

func (r *accessRepositoryImpl) CountByUserID(ctx context.Context, userID int) (int, error) {
	return r.client.Access.Query().
		Where(access.HasUserWith(user.ID(userID))).
		Count(ctx)
}
