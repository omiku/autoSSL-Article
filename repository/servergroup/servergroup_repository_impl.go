package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/servergroup"
	"autoSSL/ent/servergroupmember"
	"autoSSL/ent/user"
)

type serverGroupRepositoryImpl struct {
	client *ent.Client
}

type serverGroupMemberRepositoryImpl struct {
	client *ent.Client
}

func NewServerGroupRepository(client *ent.Client) ServerGroupRepository {
	return &serverGroupRepositoryImpl{client: client}
}

func NewServerGroupMemberRepository(client *ent.Client) ServerGroupMemberRepository {
	return &serverGroupMemberRepositoryImpl{client: client}
}

func (r *serverGroupRepositoryImpl) Create(ctx context.Context, name, description, status string, accessType string, userID int) (*ent.ServerGroup, error) {
	return r.client.ServerGroup.Create().
		SetName(name).
		SetDescription(description).
		SetType(servergroup.Type(accessType)).
		SetStatus(servergroup.Status(status)).
		SetUserID(userID).
		Save(ctx)
}

func (r *serverGroupRepositoryImpl) GetByID(ctx context.Context, userID, id int) (*ent.ServerGroup, error) {
	return r.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.ID(userID))).
		Where(servergroup.ID(id)).
		First(ctx)
}

func (r *serverGroupRepositoryImpl) GetByIDWithMembers(ctx context.Context, id int) (*ent.ServerGroup, error) {
	return r.client.ServerGroup.Query().
		Where(servergroup.ID(id)).
		WithGroupMemberships(func(q *ent.ServerGroupMemberQuery) {
			q.WithAccess()
		}).
		First(ctx)
}

func (r *serverGroupRepositoryImpl) ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.ServerGroup, error) {
	return r.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.ID(userID))).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
}

func (r *serverGroupRepositoryImpl) ListWithAccessCount(ctx context.Context, userID int, page, pageSize int) ([]*ent.ServerGroup, int, error) {
	total, err := r.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.ID(userID))).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	groups, err := r.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.ID(userID))).
		WithGroupMemberships(func(q *ent.ServerGroupMemberQuery) {
			q.WithAccess()
		}).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *serverGroupRepositoryImpl) Update(ctx context.Context, id int, userID int, name, description, status string) error {
	_, err := r.client.ServerGroup.UpdateOneID(id).
		Where(servergroup.HasUserWith(user.ID(userID))).
		SetName(name).
		SetDescription(description).
		SetStatus(servergroup.Status(status)).
		Save(ctx)
	return err
}

func (r *serverGroupRepositoryImpl) UpdateWithMembers(ctx context.Context, id int, userID int, name, description, status string, toAdd []servergroupmember.Type, toDelete []int) (*ent.ServerGroup, error) {
	_, err := r.client.ServerGroup.UpdateOneID(id).
		Where(servergroup.HasUserWith(user.ID(userID))).
		SetName(name).
		SetDescription(description).
		SetStatus(servergroup.Status(status)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.client.ServerGroup.Query().
		Where(servergroup.ID(id)).
		WithGroupMemberships(func(q *ent.ServerGroupMemberQuery) {
			q.WithAccess()
		}).
		First(ctx)
}

func (r *serverGroupRepositoryImpl) Delete(ctx context.Context, id int, userID int) error {
	return r.client.ServerGroup.DeleteOneID(id).
		Where(servergroup.HasUserWith(user.ID(userID))).
		Exec(ctx)
}

func (r *serverGroupRepositoryImpl) DeleteMembersByServerGroupID(ctx context.Context, id int) error {
	_, err := r.client.ServerGroupMember.Delete().
		Where(servergroupmember.HasServerGroupWith(servergroup.ID(id))).
		Exec(ctx)
	return err
}

func (r *serverGroupRepositoryImpl) CreateMembers(ctx context.Context, serverGroupID int, members []*ent.ServerGroupMemberCreate) error {
	if len(members) == 0 {
		return nil
	}
	_, err := r.client.ServerGroupMember.CreateBulk(members...).Save(ctx)
	return err
}

func (r *serverGroupRepositoryImpl) GetMembersByServerGroupID(ctx context.Context, id int) ([]*ent.ServerGroupMember, error) {
	return r.client.ServerGroupMember.Query().
		Where(servergroupmember.HasServerGroupWith(servergroup.ID(id))).
		WithAccess().
		All(ctx)
}

func (r *serverGroupRepositoryImpl) CountByUserID(ctx context.Context, userID int) (int, error) {
	return r.client.ServerGroup.Query().
		Where(servergroup.HasUserWith(user.ID(userID))).
		Count(ctx)
}

func (r *serverGroupMemberRepositoryImpl) Create(ctx context.Context, serverGroupID int, accessID int, memberType servergroupmember.Type) (*ent.ServerGroupMember, error) {
	return r.client.ServerGroupMember.Create().
		SetServerGroupID(serverGroupID).
		SetAccessID(accessID).
		SetType(memberType).
		Save(ctx)
}

func (r *serverGroupMemberRepositoryImpl) CreateBulk(ctx context.Context, members []*ent.ServerGroupMemberCreate) ([]*ent.ServerGroupMember, error) {
	if len(members) == 0 {
		return []*ent.ServerGroupMember{}, nil
	}
	return r.client.ServerGroupMember.CreateBulk(members...).Save(ctx)
}

func (r *serverGroupMemberRepositoryImpl) DeleteByID(ctx context.Context, id int) error {
	return r.client.ServerGroupMember.DeleteOneID(id).Exec(ctx)
}

func (r *serverGroupMemberRepositoryImpl) DeleteByIDs(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.ServerGroupMember.Delete().
		Where(servergroupmember.IDIn(ids...)).
		Exec(ctx)
	return err
}

func (r *serverGroupMemberRepositoryImpl) DeleteByServerGroupID(ctx context.Context, serverGroupID int) error {
	_, err := r.client.ServerGroupMember.Delete().
		Where(servergroupmember.HasServerGroupWith(servergroup.ID(serverGroupID))).
		Exec(ctx)
	return err
}

func (r *serverGroupMemberRepositoryImpl) GetByServerGroupID(ctx context.Context, serverGroupID int) ([]*ent.ServerGroupMember, error) {
	return r.client.ServerGroupMember.Query().
		Where(servergroupmember.HasServerGroupWith(servergroup.ID(serverGroupID))).
		WithAccess().
		All(ctx)
}
