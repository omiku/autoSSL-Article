package workflow

import (
	"autoSSL/bootstrap"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 注册工作流相关路由
func SetupRoutes(router *gin.RouterGroup, container *bootstrap.Container) {
	workflowService := container.WorkflowManager
	controller := NewController(workflowService)
	// 工作流API路由组，应用认证中间件
	workflowRoutes := router.Group("/workflows")
	{
		// 获取工作流列表
		workflowRoutes.GET("", controller.ListWorkflows)
		// 获取单个工作流详情
		workflowRoutes.GET("/:id", controller.GetWorkflowByID)
		// 更新工作流状态
		workflowRoutes.PUT("/:id/status", controller.UpdateWorkflowStatus)
		// 获取工作流步骤
		// workflowRoutes.GET("/:id/steps", controller.GetWorkflowSteps)
		// 重试工作流
		workflowRoutes.POST("/:id/retry", controller.RetryWorkflow)
		// 启动工作流
		workflowRoutes.POST("/:id/start", controller.StartWorkflow)
		// 取消工作流
		workflowRoutes.POST("/:id/cancel", controller.CancelWorkflow)
	}
}
