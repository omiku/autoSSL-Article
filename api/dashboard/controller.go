package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"autoSSL/api/response"
	"autoSSL/service/dashboard"
)

// Controller 仪表板API控制器
// @Schema 用户仪表板控制器
// @Description 提供用户登录后的监控看板数据展示接口
// @Tag 仪表板
// @BasePath /api/v1
// @Produces json
// @Consumes json
type Controller struct {
	dashboardService *dashboard.Service
}

// NewController 创建仪表板控制器实例
func NewController(dashboardService *dashboard.Service) *Controller {
	return &Controller{dashboardService: dashboardService}
}

// GetCountStats 获取数量统计数据
// @Summary 获取用户数量统计
// @Description 获取当前用户的总证书数、access数量、服务组数量、DNS供应商数量
// @Tags 仪表板
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dashboard.CountStats} "数量统计数据"
// @Failure 500 {object} response.Response "获取统计数据失败"
// @Router /dashboard/count-stats [get]
func (c *Controller) GetCountStats(ctx *gin.Context) {
	userID := ctx.GetInt("userID")

	stats, err := c.dashboardService.GetCountStats(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取统计数据失败: "+err.Error())
		return
	}

	response.Success(ctx, stats)
}

// GetRecentDeployments 获取最近部署状态
// @Summary 获取最近部署状态
// @Description 获取最近5条部署状态，包含证书名、部署类型、更新时间、状态
// @Tags 仪表板
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dashboard.DeploymentInfo} "最近部署列表"
// @Failure 500 {object} response.Response "获取部署状态失败"
// @Router /dashboard/recent-deployments [get]
func (c *Controller) GetRecentDeployments(ctx *gin.Context) {
	userID := ctx.GetInt("userID")

	deployments, err := c.dashboardService.GetRecentDeployments(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取部署状态失败: "+err.Error())
		return
	}

	response.Success(ctx, deployments)
}

// GetPendingWorkflows 获取待开始工作流
// @Summary 获取待开始工作流
// @Description 获取最近5条待开始工作流，包含证书名、部署类型、创建时间
// @Tags 仪表板
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dashboard.PendingWorkflow} "待开始工作流列表"
// @Failure 500 {object} response.Response "获取待开始工作流失败"
// @Router /dashboard/pending-workflows [get]
func (c *Controller) GetPendingWorkflows(ctx *gin.Context) {
	userID := ctx.GetInt("userID")

	workflows, err := c.dashboardService.GetPendingWorkflows(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取待开始工作流失败: "+err.Error())
		return
	}

	response.Success(ctx, workflows)
}
