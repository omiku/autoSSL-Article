package access

import (
	"autoSSL/ent"
	accessRepo "autoSSL/repository/access"
	"context"
)

type AccessService struct {
	client     *ent.Client
	accessRepo accessRepo.AccessRepository
}

func NewAccessService(client *ent.Client, accessRepo accessRepo.AccessRepository) *AccessService {
	return &AccessService{client: client, accessRepo: accessRepo}
}

// Create 创建访问配置
func (s *AccessService) Create(ctx context.Context, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
	return s.accessRepo.Create(ctx, userID, name, config, accessType, isActive)
}

// CreateV2 创建访问配置（带服务器组关联）
// func (s *AccessService) CreateV2(ctx context.Context, userID int, serverGroupID *int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
// 	create := s.client.Access.Create().
// 		SetName(name).
// 		SetConfig(config).
// 		SetType(access.Type(accessType)).
// 		SetIsActive(isActive).
// 		SetUserID(userID)

// 	if serverGroupID != nil {
// 		create.SetServerGroupID(*serverGroupID)
// 	}

// 	return create.Save(ctx)
// }

// GetByID 根据ID获取当前用户的访问配置，包含服务组信息
func (s *AccessService) GetByID(ctx context.Context, userID, id int) (*ent.Access, error) {
	return s.accessRepo.GetByID(ctx, userID, id)
}

// ListByUserID 获取用户的所有访问配置
func (s *AccessService) ListByUserID(ctx context.Context, userID int, page, pageSize int) ([]*ent.Access, error) {
	return s.accessRepo.ListByUserID(ctx, userID, page, pageSize)
}

// Update 更新访问配置
func (s *AccessService) Update(ctx context.Context, id int, userID int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
	return s.accessRepo.Update(ctx, id, userID, name, config, accessType, isActive)
}

// UpdateV2 更新访问配置（带服务器组关联）
// func (s *AccessService) UpdateV2(ctx context.Context, id int, userID int, serverGroupID *int, name string, config map[string]interface{}, accessType string, isActive bool) (*ent.Access, error) {
// 	update := s.client.Access.UpdateOneID(id).
// 		Where(access.HasUserWith(user.ID(userID))).
// 		SetName(name).
// 		SetConfig(config).
// 		SetType(access.Type(accessType)).
// 		SetIsActive(isActive)

// 	if serverGroupID != nil {
// 		update.SetServerGroupID(*serverGroupID)
// 	} else {
// 		update.ClearServerGroup()
// 	}

// 	return update.Save(ctx)
// }

// Delete 删除访问配置
func (s *AccessService) Delete(ctx context.Context, id int, userID int) error {
	return s.accessRepo.Delete(ctx, id, userID)
}
