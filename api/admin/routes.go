package admin

import (
	"autoSSL/bootstrap"
	"autoSSL/service/auth"
	"autoSSL/service/global_notification"
	"autoSSL/utils/jwt"

	"github.com/gin-gonic/gin"
)

func getAuthUserService() *auth.UserService {
	// 实际实现中需要正确初始化 UserService
	return &auth.UserService{}
}

func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container, jwtUtil *jwt.JWTManager) {
	userService := container.AuthService
	globalNotificationService := global_notification.NewService(container.DB)
	adminGroup := NewAdminController(userService, container.RBACService, globalNotificationService)

	r.GET("/users", adminGroup.GetAllUsers)
	r.PUT("/users/role", adminGroup.UpdateUserRole)

	// 全局通知配置管理 - 迁移到管理员路由下统一管理
	globalConfigGroup := r.Group("/global-config")
	{
		globalConfigGroup.GET("/email", adminGroup.AdminGetGlobalEmailConfig)
		globalConfigGroup.POST("/email", adminGroup.AdminSetGlobalEmailConfig)
		globalConfigGroup.GET("/sms", adminGroup.AdminGetGlobalSMSConfig)
		globalConfigGroup.POST("/sms", adminGroup.AdminSetGlobalSMSConfig)
	}

	// 注册RBAC相关路由
	RegisterRBACRoutes(r, adminGroup, jwtUtil)
}

// RegisterRBACRoutes 注册RBAC相关路由
func RegisterRBACRoutes(r *gin.RouterGroup, controller *AdminController, jwtUtil *jwt.JWTManager) {
	// RBAC路由组
	rbacGroup := r.Group("/rbac")

	{
		// 权限检查 (任何已认证用户都可以检查自己的权限)
		// rbacGroup.GET("/enforce", controller.Enforce)

		// 用户角色管理 (需要管理员权限)
		userRolesGroup := rbacGroup.Group("/users/roles")
		{
			userRolesGroup.GET("/:userID", controller.GetUserRoles)
			userRolesGroup.POST("", controller.AddRoleForUser)
			userRolesGroup.DELETE("", controller.RemoveRoleFromUser)
		}

		// 用户权限管理 (需要管理员权限)
		userPermissionsGroup := rbacGroup.Group("/users/permissions")
		{
			userPermissionsGroup.GET("/:userID", controller.GetUserPermissions)
		}

		// 角色管理 (需要管理员权限)
		rolesGroup := rbacGroup.Group("/roles")
		{
			rolesGroup.GET("", controller.GetAllRoles)
			rolesGroup.POST("", controller.CreateRole)
		}

		// 权限管理 (需要管理员权限)
		permissionsGroup := rbacGroup.Group("/permissions")
		{
			permissionsGroup.GET("", controller.GetAllPermissions)
			permissionsGroup.POST("", controller.CreatePermission)
		}

		// 角色权限管理 (需要管理员权限)
		rolePermissionsGroup := rbacGroup.Group("/roles/permissions")
		{
			rolePermissionsGroup.POST("", controller.AssignPermissionToRole)
			rolePermissionsGroup.DELETE("", controller.RemovePermissionFromRole)
		}
	}
}
