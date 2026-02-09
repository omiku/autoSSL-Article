package auth

import (
	"autoSSL/ent"
	"autoSSL/ent/user"
	"autoSSL/logger"
	"autoSSL/service/captcha"
	"autoSSL/utils/security"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type UserService struct {
	client *ent.Client
}

type Service struct {
	db      *ent.Client
	captcha *captcha.Service
	rbac    RBACService
}

// RBACService 定义RBAC服务接口
type RBACService interface {
	AddRoleForUser(ctx context.Context, userID int, roleName string) error
	RemoveRoleFromUser(ctx context.Context, userID int, roleName string) error
	GetUserRoles(ctx context.Context, userID int) ([]*ent.Role, error)
	ReloadPolicies() error
}

func New(db *ent.Client, captcha *captcha.Service, rbac RBACService) *Service {
	return &Service{
		db:      db,
		captcha: captcha,
		rbac:    rbac,
	}
}

// NewUserService 创建UserService实例
//	@param	client	*ent.Client	-	ent数据库客户端
//	@return	*UserService - 用户服务实例
// func NewUserService(client *ent.Client) *UserService {
// 	return &UserService{client: client}
// }

// GetUserByEmail 根据邮箱获取用户
//
//	@Summary		根据邮箱获取用户信息
//	@Description	通过用户邮箱查询用户详细信息
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return s.db.User.Query().Where(user.Email(email)).First(ctx)
}

// GetAllUsers 获取所有用户
//
//	@Summary		获取所有用户信息
//	@Description	查询系统中所有用户的详细信息
func (s *Service) GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	return s.db.User.Query().All(ctx)
}

// UpdateUserRole 更新用户角色
//
//	@Summary		更新用户角色
//	@Description	根据用户ID和新角色更新用户权限（通过RBAC服务更新关联表）
func (s *Service) UpdateUserRole(ctx context.Context, userID int, role string) error {
	// 获取用户当前角色
	currentRoles, err := s.rbac.GetUserRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户当前角色失败: %w", err)
	}

	// 移除所有现有角色
	for _, r := range currentRoles {
		if err := s.rbac.RemoveRoleFromUser(ctx, userID, r.Name); err != nil {
			return fmt.Errorf("移除用户角色失败: %w", err)
		}
	}

	// 添加新角色
	if err := s.rbac.AddRoleForUser(ctx, userID, role); err != nil {
		return fmt.Errorf("添加新角色失败: %w", err)
	}

	// 角色更新成功后，刷新RBAC权限缓存
	if err := s.rbac.ReloadPolicies(); err != nil {
		logger.Error("刷新RBAC权限缓存失败", zap.Error(err))
	} else {
		logger.Info("RBAC权限缓存已刷新", zap.Int("userID", userID), zap.String("newRole", role))
	}

	return nil
}

// GetUserRoles 获取用户角色列表
func (s *Service) GetUserRoles(ctx context.Context, userID int) ([]*ent.Role, error) {
	return s.rbac.GetUserRoles(ctx, userID)
}

// CreateUser 创建新用户
//
//	@Summary		创建新用户
//	@Description	根据提供的邮箱、密码创建新用户（第一个用户自动设为管理员）
func (s *Service) CreateUser(ctx context.Context, email string, password string, _ string) (*ent.User, error) {
	// 查询当前用户总数
	count, err := s.db.User.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询用户数量失败: %v", err)
	}

	// 第一个用户设为管理员，其余为用户
	role := "user"
	isAdmin := false
	if count == 0 {
		role = "admin"
		isAdmin = true
	}

	hashed, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user, err := s.db.User.Create().
		SetEmail(email).
		SetPasswordHash(string(hashed)).
		SetIsAdmin(isAdmin).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 通过RBAC服务为用户分配角色
	if err := s.rbac.AddRoleForUser(ctx, user.ID, role); err != nil {
		// 如果角色分配失败，记录错误但不中断注册流程
		// 用户仍然可以通过管理员手动分配角色
		logger.Error("为用户分配角色失败", zap.Error(err), zap.Int("userID", user.ID), zap.String("role", role))
	} else {
		// 角色分配成功后，刷新RBAC权限缓存
		if err := s.rbac.ReloadPolicies(); err != nil {
			logger.Error("刷新RBAC权限缓存失败", zap.Error(err))
		} else {
			logger.Info("RBAC权限缓存已刷新", zap.Int("userID", user.ID), zap.String("role", role))
		}
	}

	return user, nil
}

// GenerateCaptcha 生成验证码
//
//	@Summary		生成验证码
//	@Description	生成图片验证码并返回ID和Base64编码的图片
func (s *Service) GenerateCaptcha() (string, captcha.Base64Adapter, error) {
	return s.captcha.Generate()
}

// VerifyCaptcha 验证验证码
//
//	@Summary		验证验证码
//	@Description	根据验证码ID和用户输入的答案验证是否正确
func (s *Service) VerifyCaptcha(id, answer string) bool {
	return s.captcha.Verify(id, answer)
}
