package bootstrap

import (
	"autoSSL/config"
	"autoSSL/database"
	"autoSSL/logger"
	"context"
	"errors"
	"fmt"

	"autoSSL/ent"
	"autoSSL/ent/notificationtemplate"
	notificationRepo "autoSSL/repository/notification"
	"autoSSL/service/notification"
	"autoSSL/service/rbac"

	"go.uber.org/zap"
)

// Init 初始化所有应用依赖
// configPath 配置文件路径 (如 "./config.yaml")
func Init(configPath string) (*Container, error) {
	// 1. 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, errors.New("加载配置失败: " + err.Error())
	}

	// 2. 初始化日志
	log, err := logger.InitLogger(&cfg.Logger)
	if err != nil {
		log.Error("日志初始化失败", zap.Error(err))
		return nil, err
	}

	// 3. 初始化数据库
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Error("数据库初始化失败", zap.Error(err))
		return nil, err
	}

	// 自动迁移数据库表结构（ent）
	ctx := context.Background()
	if err = db.Schema.Create(ctx); err != nil {
		return nil, errors.New("数据库迁移失败: " + err.Error())
	}

	// 4. 初始化容器并启动服务
	container := &Container{
		Config: cfg,
		Logger: log,
		DB:     db,
	}

	// 初始化所有服务依赖
	if err := container.Init(); err != nil {
		log.Error("容器服务初始化失败", zap.Error(err))
		return nil, err
	}

	// 初始化默认角色和权限
	if err := initDefaultRolesAndPermissions(ctx, db, log); err != nil {
		log.Error("初始化默认角色和权限失败", zap.Error(err))
		return nil, err
	}

	// 初始化系统默认通知模板
	if err := initSystemTemplates(ctx, db, log); err != nil {
		log.Error("初始化系统默认通知模板失败", zap.Error(err))
		return nil, err
	}

	return container, nil
}

// initDefaultRolesAndPermissions 初始化默认角色和权限
func initDefaultRolesAndPermissions(ctx context.Context, client *ent.Client, log *logger.Logger) error {
	// 检查是否已存在角色数据
	count, err := client.Role.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("查询角色数量失败: %w", err)
	}

	// 如果已有数据，则跳过初始化
	if count > 0 {
		log.Info("已存在角色数据，跳过默认角色和权限初始化")
		return nil
	}

	// 使用事务确保数据一致性
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 创建RBAC服务实例（使用事务客户端）
	rbacService, err := rbac.NewService(tx.Client())
	if err != nil {
		return fmt.Errorf("创建RBAC服务失败: %w", err)
	}

	// 定义默认角色
	roles := []struct {
		name        string
		description string
	}{
		{"admin", "系统管理员"},
		{"user", "普通用户"},
		{"viewer", "只读用户"},
	}

	// 创建默认角色并存储引用
	createdRoles := make(map[string]*ent.Role)
	for _, r := range roles {
		role, err := rbacService.CreateRole(ctx, r.name, r.description)
		if err != nil {
			return fmt.Errorf("创建角色 %s 失败: %w", r.name, err)
		}
		createdRoles[r.name] = role
		log.Info(fmt.Sprintf("创建角色 %s 成功", r.name))
	}

	// 定义所有权限
	allPermissions := []struct {
		action   string
		resource string
		desc     string
	}{
		{"POST", "/apiv1/cert/certificates", "创建证书"},
		{"GET", "/apiv1/cert/certificates/:id", "查看证书"},
		{"GET", "/apiv1/cert/certificates", "列出证书"},
		{"GET", "/apiv1/cert/certificates/:id/download", "下载证书"},
		{"POST", "/apiv1/dns-provider/dnsproviders", "创建DNS提供商"},
		{"GET", "/apiv1/dns-provider/dnsproviders/:id", "查看DNS提供商"},
		{"GET", "/apiv1/dns-provider/dnsproviders", "列出DNS提供商"},
		{"PUT", "/apiv1/dns-provider/dnsproviders/:id", "更新DNS提供商"},
		{"DELETE", "/apiv1/dns-provider/dnsproviders/:id", "删除DNS提供商"},
		{"POST", "/apiv1/server-group/servers", "创建服务器组"},
		{"GET", "/apiv1/server-group/servers", "查看服务器组"},
		{"PUT", "/apiv1/server-group/servers/:id", "更新服务器组"},
		{"DELETE", "/apiv1/server-group/servers/:id", "删除服务器组"},
		{"POST", "/apiv1/workflow/workflows", "创建工作流"},
		{"GET", "/apiv1/workflow/workflows", "列出工作流"},
		{"GET", "/apiv1/workflow/workflows/:id", "查看工作流"},
		{"PUT", "/apiv1/workflow/workflows/:id/status", "更新工作流状态"},
		{"POST", "/apiv1/workflow/workflows/:id/retry", "重试工作流"},
		{"POST", "/apiv1/workflow/workflows/:id/start", "启动工作流"},
		{"POST", "/apiv1/workflow/workflows/:id/cancel", "取消工作流"},
		{"POST", "/apiv1/access/access", "创建访问配置"},
		{"GET", "/apiv1/access/access", "列出访问配置"},
		{"GET", "/apiv1/access/access/:id", "查看访问配置"},
		{"PUT", "/apiv1/access/access/:id", "更新访问配置"},
		{"DELETE", "/apiv1/access/access/:id", "删除访问配置"},
		{"GET", "/apiv1/admin/users", "查看用户列表"},
		{"PUT", "/apiv1/admin/users/role", "更新用户角色"},
		{"GET", "/apiv1/admin/rbac/users/roles/:userID", "查看用户角色"},
		{"POST", "/apiv1/admin/rbac/users/roles", "为用户添加角色"},
		{"DELETE", "/apiv1/admin/rbac/users/roles", "为用户移除角色"},
		{"GET", "/apiv1/admin/rbac/users/permissions/:userID", "查看用户权限"},
		{"GET", "/apiv1/admin/rbac/roles", "查看角色列表"},
		{"POST", "/apiv1/admin/rbac/roles", "创建角色"},
		{"GET", "/apiv1/admin/rbac/permissions", "查看权限列表"},
		{"POST", "/apiv1/admin/rbac/permissions", "创建权限"},
		{"POST", "/apiv1/admin/rbac/roles/permissions", "为角色分配权限"},
		{"DELETE", "/apiv1/admin/rbac/roles/permissions", "为角色移除权限"},
		{"GET", "/apiv1/admin/global-config/email", "管理员查看全局邮件配置"},
		{"POST", "/apiv1/admin/global-config/email", "管理员设置全局邮件配置"},
		{"GET", "/apiv1/admin/global-config/sms", "管理员查看全局短信配置"},
		{"POST", "/apiv1/admin/global-config/sms", "管理员设置全局短信配置"},
		{"GET", "/apiv1/dashboard/count-stats", "面板统计信息"},
		{"GET", "/apiv1/dashboard/pending-workflows", "待开始的工作流"},
		{"GET", "/apiv1/dashboard/recent-deployments", "最近的部署"},
		{"GET", "/apiv1/notification/config", "查看通知配置"},
		{"POST", "/apiv1/notification/config", "创建通知配置"},
		{"GET", "/apiv1/notification/config/:id", "查看指定通知配置"},
		{"PUT", "/apiv1/notification/config/:id", "更新通知配置"},
		{"DELETE", "/apiv1/notification/config/:id", "删除通知配置"},
		{"GET", "/apiv1/notification/global/email", "查看全局邮件配置"},
		{"GET", "/apiv1/notification/global/sms", "查看全局短信配置"},
		{"GET", "/apiv1/notification/templates", "查看通知模板列表"},
		{"POST", "/apiv1/notification/templates", "创建通知模板"},
		{"GET", "/apiv1/notification/templates/system", "查看系统通知模板"},
		{"GET", "/apiv1/notification/templates/options", "查看通知模板选项"},
		{"GET", "/apiv1/notification/templates/:id", "查看指定通知模板"},
		{"PUT", "/apiv1/notification/templates/:id", "更新通知模板"},
		{"DELETE", "/apiv1/notification/templates/:id", "删除通知模板"},
	}

	// 创建所有权限并建立关联
	for _, p := range allPermissions {
		// 创建权限
		perm, err := tx.Permission.Create().
			SetResource(p.resource).
			SetAction(p.action).
			SetDescription(p.desc).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("创建权限 %s:%s 失败: %w", p.resource, p.action, err)
		}

		// 为admin角色分配所有权限
		if err := createdRoles["admin"].Update().AddPermissions(perm).Exec(ctx); err != nil {
			return fmt.Errorf("为admin角色分配权限 %s:%s 失败: %w", p.resource, p.action, err)
		}

		// 为user角色分配特定权限
		isUserPermission := false
		for _, up := range []string{
			"/apiv1/cert/certificates:POST", "/apiv1/cert/certificates:GET", "/apiv1/cert/certificates/:id:GET", "/apiv1/cert/certificates/:id/download:GET",
			"/apiv1/dns-provider/dnsproviders:POST", "/apiv1/dns-provider/dnsproviders:GET", "/apiv1/dns-provider/dnsproviders/:id:GET", "/apiv1/dns-provider/dnsproviders/:id:PUT", "/apiv1/dns-provider/dnsproviders/:id:DELETE",
			"/apiv1/server-group/servers:POST", "/apiv1/server-group/servers:GET", "/apiv1/server-group/servers/:id:PUT", "/apiv1/server-group/servers/:id:DELETE",
			"/apiv1/workflow/workflows:POST", "/apiv1/workflow/workflows:GET", "/apiv1/workflow/workflows/:id:GET", "/apiv1/workflow/workflows/:id/status:PUT", "/apiv1/workflow/workflows/:id/retry:POST", "/apiv1/workflow/workflows/:id/start:POST", "/apiv1/workflow/workflows/:id/cancel:POST",
			"/apiv1/access/access:POST", "/apiv1/access/access:GET", "/apiv1/access/access/:id:GET", "/apiv1/access/access/:id:PUT", "/apiv1/access/access/:id:DELETE", "/apiv1/dashboard/count-stats", "/apiv1/dashboard/pending-workflows", "/apiv1/dashboard/recent-deployments",
			"/apiv1/notification/config:GET", "/apiv1/notification/config:POST", "/apiv1/notification/config/:id:GET", "/apiv1/notification/config/:id:PUT", "/apiv1/notification/config/:id:DELETE", "/apiv1/notification/global/email:GET", "/apiv1/notification/global/sms:GET",
			"/apiv1/notification/templates:GET", "/apiv1/notification/templates:POST", "/apiv1/notification/templates/:id:GET", "/apiv1/notification/templates/:id:PUT", "/apiv1/notification/templates/:id:DELETE", "/apiv1/notification/templates/system:GET", "/apiv1/notification/templates/options:GET",
		} {
			if fmt.Sprintf("%s:%s", p.resource, p.action) == up {
				isUserPermission = true
				break
			}
		}
		if isUserPermission {
			if err := createdRoles["user"].Update().AddPermissions(perm).Exec(ctx); err != nil {
				return fmt.Errorf("为user角色分配权限 %s:%s 失败: %w", p.resource, p.action, err)
			}
		}

		// 为viewer角色分配只读权限
		isViewerPermission := false
		for _, vp := range []string{
			"/apiv1/cert/certificates:GET", "/apiv1/cert/certificates/:id:GET", "/apiv1/cert/certificates/:id/download:GET",
			"/apiv1/dns-provider/dnsproviders:GET", "/apiv1/dns-provider/dnsproviders/:id:GET",
			"/apiv1/server-group/servers:GET", "/apiv1/workflow/workflows:GET", "/apiv1/workflow/workflows/:id:GET",
			"/apiv1/access/access:GET", "/apiv1/access/access/:id:GET",
			"/apiv1/dashboard/count-stats", "/apiv1/dashboard/pending-workflows", "/apiv1/dashboard/recent-deployments",
			"/apiv1/notification/config:GET", "/apiv1/notification/config/:id:GET", "/apiv1/notification/global/email:GET", "/apiv1/notification/global/sms:GET",
			"/apiv1/notification/templates:GET", "/apiv1/notification/templates/:id:GET", "/apiv1/notification/templates/system:GET", "/apiv1/notification/templates/options:GET",
		} {
			if fmt.Sprintf("%s:%s", p.resource, p.action) == vp {
				isViewerPermission = true
				break
			}
		}
		if isViewerPermission {
			if err := createdRoles["viewer"].Update().AddPermissions(perm).Exec(ctx); err != nil {
				return fmt.Errorf("为viewer角色分配权限 %s:%s 失败: %w", p.resource, p.action, err)
			}
		}

		log.Info(fmt.Sprintf("创建权限 %s:%s 并分配完成", p.resource, p.action))
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Info("默认角色和权限初始化完成")
	return nil
}

// initSystemTemplates 初始化系统默认通知模板
func initSystemTemplates(ctx context.Context, client *ent.Client, log *logger.Logger) error {
	// 检查是否已存在系统模板
	count, err := client.NotificationTemplate.Query().
		Where(notificationtemplate.IsSystemDefault(true)).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("查询系统模板数量失败: %w", err)
	}

	// 如果已有系统模板，则跳过初始化
	if count > 0 {
		log.Info("已存在系统通知模板，跳过初始化")
		return nil
	}

	// 初始化模板服务
	notificationTemplateRepo := notificationRepo.NewNotificationTemplateRepository(client)
	templateService := notification.NewTemplateService(client, notificationTemplateRepo)

	// 初始化系统默认模板
	if err := templateService.InitializeSystemTemplates(ctx); err != nil {
		return fmt.Errorf("初始化系统默认模板失败: %w", err)
	}

	log.Info("系统默认通知模板初始化完成")
	return nil
}
