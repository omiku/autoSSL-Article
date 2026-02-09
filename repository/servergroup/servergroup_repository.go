package repository

import (
	"context"

	"autoSSL/ent"
	"autoSSL/ent/servergroupmember"
)

type ServerGroupRepository interface {
	Create(ctx context.Context, name, description, status string, accessType string, userID int) (*ent.ServerGroup, error)
	GetByID(ctx context.Context, userID, id int) (*ent.ServerGroup, error)
	GetByIDWithMembers(ctx context.Context, id int) (*ent.ServerGroup, error)
	ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.ServerGroup, error)
	ListWithAccessCount(ctx context.Context, userID int, page, pageSize int) ([]*ent.ServerGroup, int, error)
	Update(ctx context.Context, id int, userID int, name, description, status string) error
	UpdateWithMembers(ctx context.Context, id int, userID int, name, description, status string, toAdd []servergroupmember.Type, toDelete []int) (*ent.ServerGroup, error)
	Delete(ctx context.Context, id int, userID int) error
	DeleteMembersByServerGroupID(ctx context.Context, id int) error
	CreateMembers(ctx context.Context, serverGroupID int, members []*ent.ServerGroupMemberCreate) error
	GetMembersByServerGroupID(ctx context.Context, id int) ([]*ent.ServerGroupMember, error)
	CountByUserID(ctx context.Context, userID int) (int, error)
}

type ServerGroupMemberRepository interface {
	Create(ctx context.Context, serverGroupID int, accessID int, memberType servergroupmember.Type) (*ent.ServerGroupMember, error)
	CreateBulk(ctx context.Context, members []*ent.ServerGroupMemberCreate) ([]*ent.ServerGroupMember, error)
	DeleteByID(ctx context.Context, id int) error
	DeleteByIDs(ctx context.Context, ids []int) error
	DeleteByServerGroupID(ctx context.Context, serverGroupID int) error
	GetByServerGroupID(ctx context.Context, serverGroupID int) ([]*ent.ServerGroupMember, error)
}
