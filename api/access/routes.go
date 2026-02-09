package access

import (
	"autoSSL/bootstrap"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册访问配置相关路由
func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container) {
	// 创建访问服务实例
	accessService := container.AccessService
	controller := NewAccessController(accessService)

	// 访问配置路由组
	// V1版本接口（保持向后兼容）
	r.POST("access", controller.Create)
	r.GET("access", controller.List)
	r.GET("access/:id", controller.Get)
	r.PUT("access/:id", controller.Update)
	r.DELETE("access/:id", controller.Delete)

}
