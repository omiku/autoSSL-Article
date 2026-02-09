package dashboard

import (
	"autoSSL/bootstrap"
	"autoSSL/ent"
	"autoSSL/service/dashboard"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置仪表板路由
func SetupRoutes(group *gin.RouterGroup, container *bootstrap.Container) {
	// 创建仪表板服务
	dashboardService := NewDashboardService(container.DB)
	controller := NewController(dashboardService)

	// 仪表板相关路由
	dashboard := group.Group("/dashboard")
	{
		// 数量统计数据
		dashboard.GET("/count-stats", controller.GetCountStats)

		// 最近部署状态
		dashboard.GET("/recent-deployments", controller.GetRecentDeployments)

		// 待开始工作流
		dashboard.GET("/pending-workflows", controller.GetPendingWorkflows)
	}
}

// NewDashboardService 创建仪表板服务实例
func NewDashboardService(db *ent.Client) *dashboard.Service {
	return dashboard.NewService(db)
}
