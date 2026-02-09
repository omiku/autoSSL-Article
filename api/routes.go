package api

import (
	"autoSSL/api/access"
	"autoSSL/api/admin"
	"autoSSL/api/auth"
	"autoSSL/api/certificate"
	"autoSSL/api/dashboard"
	"autoSSL/api/dnsprovider"

	"autoSSL/api/notification"
	"autoSSL/api/servergroup"
	"autoSSL/api/workflow"
	"autoSSL/bootstrap"

	"github.com/gin-contrib/cors"

	_ "autoSSL/docs" // 导入Swagger文档生成的包

	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, container *bootstrap.Container) {
	// 配置CORS中间件解决跨域问题
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有源（生产环境建议指定具体域名）
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// v1路由
	v1 := r.Group("/apiv1")
	// 无需鉴权的路由组
	// v1子路由
	{
		publicGroup := v1.Group("")
		{
			// 认证路由（注入Container）
			authGroup := publicGroup.Group("/auth")
			auth.SetupRoutes(authGroup, container)
		}
		// 需要鉴权的路由组（注入JWT中间件）
		privateGroup := v1.Group("")
		privateGroup.Use(auth.JWTMiddleware(container.JWTUtil), auth.RoleMiddleware(container.RBACService)) // 使用Container中的JWT工具和RBAC服务
		{
			// 证书管理路由（注入Container）
			certGroup := privateGroup.Group("/cert")
			certificate.SetupRoutes(certGroup, container)

			// 服务器组路由（注入Container）
			serverGroup := privateGroup.Group("/server-group")
			servergroup.SetupRoutes(serverGroup, container)

			// 访问管理路由（注入Container）
			accessGroup := privateGroup.Group("/access")
			access.SetupRoutes(accessGroup, container)

			// DNS提供商路由（注入Container）
			dnsProviderGroup := privateGroup.Group("/dns-provider")
			dnsprovider.SetupRoutes(dnsProviderGroup, container)

			// 管理员组路由（注入Container）
			adminGroup := privateGroup.Group("/admin")
			admin.SetupRoutes(adminGroup, container, container.JWTUtil)

			// 工作流路由（注入Container）
			workflowGroup := privateGroup.Group("/workflow")
			workflow.SetupRoutes(workflowGroup, container)

			// 通知配置路由（注入Container）
			notificationGroup := privateGroup.Group("/notification")
			notification.SetupRoutes(notificationGroup, container)

			// 仪表板路由（注入Container）
			dashboardGroup := privateGroup.Group("")
			dashboard.SetupRoutes(dashboardGroup, container)
		}
	}
}
