package notification

import (
	"autoSSL/bootstrap"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置通知配置路由
func SetupRoutes(router *gin.RouterGroup, container *bootstrap.Container) {
	controller := NewController(container.DB, container.GlobalNotificationService)
	templateController := NewTemplateController(container.NotificationTemplateService)

	// 用户通知配置
	notification := router.Group("/config")
	{
		notification.GET("", controller.ListNotificationConfigs)
		notification.POST("", controller.CreateNotificationConfig)
		notification.PUT("/:id", controller.UpdateNotificationConfig)
		notification.DELETE("/:id", controller.DeleteNotificationConfig)
	}

	// 全局通知配置端点（用户只读）
	global := router.Group("/global")
	{
		global.GET("/email", controller.GetGlobalEmailConfig)
		global.GET("/sms", controller.GetGlobalSMSConfig)
	}

	// 通知模板管理
	templates := router.Group("/templates")
	{
		templates.GET("", templateController.ListUserTemplates)
		templates.POST("", templateController.CreateTemplate)
		templates.GET("/system", templateController.ListSystemTemplates)
		templates.GET("/options", templateController.TemplateTypeOptions)
		templates.GET("/:id", templateController.GetTemplate)
		templates.PUT("/:id", templateController.UpdateTemplate)
		templates.DELETE("/:id", templateController.DeleteTemplate)
	}
}
