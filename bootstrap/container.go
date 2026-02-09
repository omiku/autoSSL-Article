package bootstrap

import (
	"context"

	"autoSSL/config"
	"autoSSL/ent"
	"autoSSL/logger"
	accessRepo "autoSSL/repository/access"
	certificateRepo "autoSSL/repository/certificate"
	dnsproviderRepo "autoSSL/repository/dnsprovider"
	notificationRepo "autoSSL/repository/notification"
	servergroupRepo "autoSSL/repository/servergroup"
	wfRepo "autoSSL/repository/workflow"
	"autoSSL/service/access"
	"autoSSL/service/auth"
	"autoSSL/service/captcha"
	"autoSSL/service/certificate"
	"autoSSL/service/dnsprovider"
	"autoSSL/service/global_notification"
	"autoSSL/service/notification"
	"autoSSL/service/rbac"
	"autoSSL/service/servergroup"
	"autoSSL/service/workflow"
	"autoSSL/utils/jwt"

	"go.uber.org/zap"
)

// Container 集中管理所有需要初始化的依赖
type Container struct {
	Config *config.AppConfig // 全局配置
	Logger *logger.Logger    // 日志实例
	DB     *ent.Client       // 数据库客户端

	// Repository 层依赖
	WorkflowRepository             wfRepo.WorkflowRepository
	WorkflowStepRepository         wfRepo.WorkflowStepRepository
	CertificateRepository          certificateRepo.CertificateRepository
	AccessRepository               accessRepo.AccessRepository
	DNSProviderRepository          dnsproviderRepo.DNSProviderRepository
	ServerGroupRepository          servergroupRepo.ServerGroupRepository
	ServerGroupMemberRepository    servergroupRepo.ServerGroupMemberRepository
	NotificationConfigRepository   notificationRepo.NotificationConfigRepository
	NotificationTemplateRepository notificationRepo.NotificationTemplateRepository

	// 服务层依赖
	AuthService                 *auth.Service
	CaptchaService              *captcha.Service
	CertService                 *certificate.CertificateService
	DNSService                  *dnsprovider.Service
	AccessService               *access.AccessService
	ServerGroupSvc              *servergroup.ServerGroupService
	WorkflowManager             *workflow.Manager
	WorkflowScheduler           *workflow.Scheduler
	GlobalNotificationService   *global_notification.Service
	NotificationTemplateService *notification.TemplateService

	// 工具类依赖
	JWTUtil *jwt.JWTManager

	// RBAC服务
	RBACService *rbac.Service
}

// Init 初始化所有服务依赖
// 返回错误信息以便调用者处理初始化失败情况
func (c *Container) Init() error {
	// 初始化 Repository 层
	c.WorkflowRepository = wfRepo.NewWorkflowRepository(c.DB)
	c.WorkflowStepRepository = wfRepo.NewWorkflowStepRepository(c.DB)
	c.CertificateRepository = certificateRepo.NewCertificateRepository(c.DB)
	c.AccessRepository = accessRepo.NewAccessRepository(c.DB)
	c.DNSProviderRepository = dnsproviderRepo.NewDNSProviderRepository(c.DB)
	c.ServerGroupRepository = servergroupRepo.NewServerGroupRepository(c.DB)
	c.ServerGroupMemberRepository = servergroupRepo.NewServerGroupMemberRepository(c.DB)
	c.NotificationConfigRepository = notificationRepo.NewNotificationConfigRepository(c.DB)
	c.NotificationTemplateRepository = notificationRepo.NewNotificationTemplateRepository(c.DB)

	// 初始化工作流服务
	c.WorkflowManager = workflow.NewManager(c.DB)
	c.WorkflowScheduler = workflow.NewScheduler(c.DB, &workflow.SchedulerConfig{
		MaxConcurrentRenewals: 5,
		RenewalCheckCron:      "0 0 * * *", // 每天凌晨检查
	})
	c.WorkflowScheduler.Start()

	// 初始化其他服务
	c.CaptchaService, _ = captcha.New(c.Config.Captcha)
	c.RBACService, _ = rbac.NewService(c.DB)
	c.AuthService = auth.New(c.DB, c.CaptchaService, c.RBACService)
	c.CertService = certificate.NewCertificateService(c.DB, c.WorkflowManager, c.CertificateRepository)
	c.DNSService = dnsprovider.NewService(c.DB, c.DNSProviderRepository)
	c.AccessService = access.NewAccessService(c.DB, c.AccessRepository)
	c.ServerGroupSvc = servergroup.NewServerGroupService(c.DB, c.ServerGroupRepository, c.ServerGroupMemberRepository, c.AccessRepository)
	c.GlobalNotificationService = global_notification.NewService(c.DB)
	c.NotificationTemplateService = notification.NewTemplateService(c.DB, c.NotificationTemplateRepository)
	c.JWTUtil = jwt.New(&c.Config.JWT)

	// 初始化系统模板
	if err := c.NotificationTemplateService.InitializeSystemTemplates(context.Background()); err != nil {
		c.Logger.Error("初始化系统模板失败", zap.Error(err))
		return err
	}

	return nil
}

// Close 关闭数据库连接
func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
